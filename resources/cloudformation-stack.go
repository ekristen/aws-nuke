package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudformationMaxDeleteAttempt = 3

const CloudFormationStackResource = "CloudFormationStack"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudFormationStackResource,
		Scope:  nuke.Account,
		Lister: &CloudFormationStackLister{},
	})
}

type CloudFormationStackLister struct{}

func (l *CloudFormationStackLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudformation.New(opts.Session)

	params := &cloudformation.DescribeStacksInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeStacks(params)
		if err != nil {
			return nil, err
		}
		for _, stack := range resp.Stacks {
			resources = append(resources, &CloudFormationStack{
				svc:               svc,
				stack:             stack,
				maxDeleteAttempts: CloudformationMaxDeleteAttempt,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudFormationStack struct {
	svc               cloudformationiface.CloudFormationAPI
	stack             *cloudformation.Stack
	maxDeleteAttempts int
	settings          *settings.Setting
}

func (cfs *CloudFormationStack) Filter() error {
	if ptr.ToString(cfs.stack.Description) == "DO NOT MODIFY THIS STACK! This stack is managed by Config Conformance Packs." {
		return fmt.Errorf("stack is managed by Config Conformance Packs")
	}
	return nil
}

func (cfs *CloudFormationStack) Settings(setting *settings.Setting) {
	cfs.settings = setting
}

func (cfs *CloudFormationStack) Remove(_ context.Context) error {
	return cfs.removeWithAttempts(0)
}

func (cfs *CloudFormationStack) removeWithAttempts(attempt int) error {
	if err := cfs.doRemove(); err != nil {
		// TODO: pass logrus in via ListerOpts so that it can be used here instead of global

		logrus.Errorf("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d delete failed: %s",
			*cfs.stack.StackName, attempt, cfs.maxDeleteAttempts, err.Error())

		var awsErr awserr.Error
		ok := errors.As(err, &awsErr)
		if ok && awsErr.Code() == "ValidationError" &&
			awsErr.Message() == "Stack ["+*cfs.stack.StackName+"] cannot be deleted while TerminationProtection is enabled" {
			// check if the setting for the resource is set to allow deletion protection to be disabled
			if cfs.settings.Get("DisableDeletionProtection").(bool) {
				logrus.Infof("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d updating termination protection",
					*cfs.stack.StackName, attempt, cfs.maxDeleteAttempts)
				_, err = cfs.svc.UpdateTerminationProtection(&cloudformation.UpdateTerminationProtectionInput{
					EnableTerminationProtection: aws.Bool(false),
					StackName:                   cfs.stack.StackName,
				})
				if err != nil {
					return err
				}
			} else {
				logrus.Warnf("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d set feature flag to disable deletion protection",
					*cfs.stack.StackName, attempt, cfs.maxDeleteAttempts)
				return err
			}
		}
		if attempt >= cfs.maxDeleteAttempts {
			return errors.New("CFS might not be deleted after this run")
		} else {
			return cfs.removeWithAttempts(attempt + 1)
		}
	} else {
		return nil
	}
}

func GetParentStack(svc cloudformationiface.CloudFormationAPI, stackID string) (*cloudformation.Stack, error) {
	o, err := svc.DescribeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		return nil, err
	}

	for _, o := range o.Stacks {
		if *o.StackId == stackID {
			return o, nil
		}
	}

	return nil, nil //nolint:nilnil
}

func (cfs *CloudFormationStack) doRemove() error { //nolint:gocyclo
	if cfs.stack.ParentId != nil {
		p, err := GetParentStack(cfs.svc, *cfs.stack.ParentId)
		if err != nil {
			return err
		}

		if p != nil {
			return liberrors.ErrHoldResource("waiting for parent stack")
		}
	}

	o, err := cfs.svc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: cfs.stack.StackName,
	})
	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			if awsErr.Code() == "ValidationFailed" && strings.HasSuffix(awsErr.Message(), " does not exist") {
				logrus.Infof("CloudFormationStack stackName=%s no longer exists", *cfs.stack.StackName)
				return nil
			}
		}
		return err
	}
	stack := o.Stacks[0]

	if *stack.StackStatus == cloudformation.StackStatusDeleteComplete {
		// stack already deleted, no need to re-delete
		return nil
	} else if *stack.StackStatus == cloudformation.StackStatusDeleteInProgress {
		logrus.Infof("CloudFormationStack stackName=%s delete in progress. Waiting", *cfs.stack.StackName)
		return cfs.svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: cfs.stack.StackName,
		})
	} else if *stack.StackStatus == cloudformation.StackStatusDeleteFailed {
		logrus.Infof("CloudFormationStack stackName=%s delete failed. Attempting to retain and delete stack", *cfs.stack.StackName)
		// This means the CFS has undetectable resources.
		// In order to move on with nuking, we retain them in the deletion.
		retainableResources, err := cfs.svc.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: cfs.stack.StackName,
		})
		if err != nil {
			return err
		}

		retain := make([]*string, 0)

		for _, r := range retainableResources.StackResourceSummaries {
			if *r.ResourceStatus != cloudformation.ResourceStatusDeleteComplete {
				retain = append(retain, r.LogicalResourceId)
			}
		}

		_, err = cfs.svc.DeleteStack(&cloudformation.DeleteStackInput{
			StackName:       cfs.stack.StackName,
			RetainResources: retain,
		})
		if err != nil {
			return err
		}
		return cfs.svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: cfs.stack.StackName,
		})
	} else {
		if err := cfs.waitForStackToStabilize(*stack.StackStatus); err != nil {
			return err
		} else if _, err := cfs.svc.DeleteStack(&cloudformation.DeleteStackInput{
			StackName: cfs.stack.StackName,
		}); err != nil {
			return err
		} else if err := cfs.svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: cfs.stack.StackName,
		}); err != nil {
			return err
		} else {
			return nil
		}
	}
}
func (cfs *CloudFormationStack) waitForStackToStabilize(currentStatus string) error {
	switch currentStatus {
	case cloudformation.StackStatusUpdateInProgress,
		cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateRollbackInProgress:
		logrus.Infof("CloudFormationStack stackName=%s update in progress. Waiting to stabalize", *cfs.stack.StackName)

		return cfs.svc.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
			StackName: cfs.stack.StackName,
		})
	case cloudformation.StackStatusCreateInProgress,
		cloudformation.StackStatusRollbackInProgress:
		logrus.Infof("CloudFormationStack stackName=%s create in progress. Waiting to stabalize", *cfs.stack.StackName)

		return cfs.svc.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: cfs.stack.StackName,
		})
	default:
		return nil
	}
}

func (cfs *CloudFormationStack) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", cfs.stack.StackName)
	properties.Set("CreationTime", cfs.stack.CreationTime.Format(time.RFC3339))
	if cfs.stack.LastUpdatedTime == nil {
		properties.Set("LastUpdatedTime", cfs.stack.CreationTime.Format(time.RFC3339))
	} else {
		properties.Set("LastUpdatedTime", cfs.stack.LastUpdatedTime.Format(time.RFC3339))
	}

	for _, tagValue := range cfs.stack.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (cfs *CloudFormationStack) String() string {
	return *cfs.stack.StackName
}
