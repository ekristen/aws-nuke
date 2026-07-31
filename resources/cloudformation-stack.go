package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"                                        //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/awserr"                                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudformation"                     //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/sts"                                //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/sts/stsiface"                       //nolint:staticcheck

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudformationMaxDeleteAttempt = 3

const CloudFormationStackResource = "CloudFormationStack"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFormationStackResource,
		Scope:    nuke.Account,
		Resource: &CloudFormationStack{},
		Lister:   &CloudFormationStackLister{},
		Settings: []string{
			"DisableDeletionProtection",
			"CreateRoleToDeleteStack",
			"UseCurrentRoleToDeleteStack",
		},
	})
}

// iamRoleAPI is the subset of the IAM v2 client used by CloudFormationStack for role
// create/delete operations. Defined as an interface to enable test mocking.
type iamRoleAPI interface {
	CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
}

type CloudFormationStackLister struct{}

func (l *CloudFormationStackLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudformation.New(opts.Session)
	// Pinned to aws-global; see awsutil.SkipRegionalForGlobalService.
	iamSvc := iam.NewFromConfig(*opts.Config, func(o *iam.Options) {
		o.Region = "aws-global"
	})
	stsSvc := sts.New(opts.Session)

	params := &cloudformation.DescribeStacksInput{}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.DescribeStacks(params)
		if err != nil {
			return nil, err
		}
		for _, stack := range resp.Stacks {
			newResource := &CloudFormationStack{
				svc:               svc,
				stsSvc:            stsSvc,
				iamSvc:            iamSvc,
				logger:            opts.Logger,
				maxDeleteAttempts: CloudformationMaxDeleteAttempt,
				Name:              stack.StackName,
				Status:            stack.StackStatus,
				description:       stack.Description,
				parentID:          stack.ParentId,
				roleARN:           stack.RoleARN,
				CreationTime:      stack.CreationTime,
				LastUpdatedTime:   stack.LastUpdatedTime,
				Tags:              stack.Tags,
			}

			if newResource.LastUpdatedTime == nil {
				newResource.LastUpdatedTime = newResource.CreationTime
			}

			resources = append(resources, newResource)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudFormationStack struct {
	svc                cloudformationiface.CloudFormationAPI
	stsSvc             stsiface.STSAPI
	iamSvc             iamRoleAPI
	settings           *settings.Setting
	logger             *logrus.Entry
	Name               *string
	Status             *string
	CreationTime       *time.Time
	LastUpdatedTime    *time.Time
	Tags               []*cloudformation.Tag
	description        *string
	parentID           *string
	roleARN            *string
	callerRoleARN      *string
	callerRoleResolved bool
	maxDeleteAttempts  int
	roleCreated        bool
	roleName           string
}

func (r *CloudFormationStack) Filter() error {
	if ptr.ToString(r.description) == "DO NOT MODIFY THIS STACK! This stack is managed by Config Conformance Packs." {
		return fmt.Errorf("stack is managed by Config Conformance Packs")
	}
	return nil
}

func (r *CloudFormationStack) Settings(setting *settings.Setting) {
	r.settings = setting
}

func (r *CloudFormationStack) Remove(ctx context.Context) error {
	return r.removeWithAttempts(ctx, 0)
}

func (r *CloudFormationStack) createRole(ctx context.Context) error {
	roleParts := strings.Split(ptr.ToString(r.roleARN), "/")
	_, err := r.iamSvc.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName: ptr.String(roleParts[len(roleParts)-1]),
		AssumeRolePolicyDocument: ptr.String(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudformation.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`),
		Tags: []iamtypes.Tag{
			{
				Key:   ptr.String("Managed"),
				Value: ptr.String("aws-nuke"),
			},
		},
	})

	if err != nil {
		return err
	}

	r.roleCreated = true
	r.roleName = roleParts[len(roleParts)-1]

	return nil
}

func (r *CloudFormationStack) removeRole(ctx context.Context) error {
	if !r.roleCreated {
		return nil
	}

	_, err := r.iamSvc.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: ptr.String(r.roleName),
	})
	return err
}

// resolveCallerRoleARN lazily resolves the caller's IAM role ARN from STS on first call.
// This is only invoked when UseCurrentRoleToDeleteStack is enabled, avoiding unnecessary
// STS API calls for users who don't use this setting.
// Note: IAM roles with path prefixes (e.g. /my-path/MyRole) cannot be fully reconstructed
// from the STS assumed-role ARN because STS omits the path component.
func (r *CloudFormationStack) resolveCallerRoleARN() *string {
	if r.callerRoleResolved {
		return r.callerRoleARN
	}
	r.callerRoleResolved = true

	identity, err := r.stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		r.logger.Warnf("CloudFormationStack stackName=%s failed to resolve caller role ARN: %s", *r.Name, err.Error())
		return nil
	}
	if identity.Arn == nil {
		r.logger.Warnf("CloudFormationStack stackName=%s GetCallerIdentity returned nil ARN", *r.Name)
		return nil
	}

	// Convert assumed-role ARN (arn:<partition>:sts::<ACCT>:assumed-role/ROLE/SESSION)
	// to role ARN (arn:<partition>:iam::<ACCT>:role/ROLE)
	arnStr := *identity.Arn
	if !strings.Contains(arnStr, ":assumed-role/") {
		r.logger.Warnf("CloudFormationStack stackName=%s caller ARN is not an assumed-role ARN (%s), cannot resolve role ARN", *r.Name, arnStr)
		return nil
	}

	parts := strings.Split(arnStr, ":")
	if len(parts) < 6 {
		r.logger.Warnf("CloudFormationStack stackName=%s caller ARN has unexpected format (%s)", *r.Name, arnStr)
		return nil
	}

	partition := parts[1]
	accountID := parts[4]
	rolePart := strings.Split(parts[5], "/")
	if len(rolePart) < 2 {
		r.logger.Warnf("CloudFormationStack stackName=%s caller ARN resource segment has unexpected format (%s)", *r.Name, parts[5])
		return nil
	}

	roleARN := fmt.Sprintf("arn:%s:iam::%s:role/%s", partition, accountID, rolePart[1])
	r.callerRoleARN = &roleARN
	return r.callerRoleARN
}

// applyRoleOverride sets the RoleARN on a DeleteStackInput if UseCurrentRoleToDeleteStack is enabled.
func (r *CloudFormationStack) applyRoleOverride(input *cloudformation.DeleteStackInput) {
	if !r.settings.GetBool("UseCurrentRoleToDeleteStack") {
		return
	}
	callerRole := r.resolveCallerRoleARN()
	if callerRole != nil {
		r.logger.Infof("CloudFormationStack stackName=%s UseCurrentRoleToDeleteStack: overriding RoleARN with %s", *r.Name, *callerRole)
		input.RoleARN = callerRole
	} else {
		r.logger.Warnf("CloudFormationStack stackName=%s UseCurrentRoleToDeleteStack enabled "+
			"but callerRoleARN could not be resolved, falling back to default role behavior", *r.Name)
	}
}

func (r *CloudFormationStack) removeWithAttempts(ctx context.Context, attempt int) error {
	if err := r.doRemove(); err != nil {
		r.logger.Errorf("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d delete failed: %s",
			*r.Name, attempt, r.maxDeleteAttempts, err.Error())

		var awsErr awserr.Error
		ok := errors.As(err, &awsErr)
		if ok && awsErr.Code() == "ValidationError" {
			// roleARN could be nil. It is not mandatory to have a roleARN for a stack.
			if r.roleARN != nil && awsErr.Message() == fmt.Sprintf("Role %s is invalid or cannot be assumed", *r.roleARN) {
				if r.settings.GetBool("CreateRoleToDeleteStack") {
					r.logger.Infof("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d creating role to delete stack",
						*r.Name, attempt, r.maxDeleteAttempts)
					if err := r.createRole(ctx); err != nil {
						return err
					}
				} else {
					r.logger.Warnf("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d set feature flag to create role to delete stack",
						*r.Name, attempt, r.maxDeleteAttempts)
					return err
				}
			} else if strings.Contains(awsErr.Message(), "cannot be deleted while TerminationProtection is enabled") &&
				strings.Contains(awsErr.Message(), *r.Name) {
				// check if the setting for the resource is set to allow deletion protection to be disabled
				if r.settings.GetBool("DisableDeletionProtection") {
					r.logger.Infof("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d updating termination protection",
						*r.Name, attempt, r.maxDeleteAttempts)
					_, err = r.svc.UpdateTerminationProtection(&cloudformation.UpdateTerminationProtectionInput{
						EnableTerminationProtection: aws.Bool(false),
						StackName:                   r.Name,
					})
					if err != nil {
						return err
					}
				} else {
					r.logger.Warnf("CloudFormationStack stackName=%s attempt=%d maxAttempts=%d set feature flag to disable deletion protection",
						*r.Name, attempt, r.maxDeleteAttempts)
					return err
				}
			}
		}

		if attempt >= r.maxDeleteAttempts {
			return errors.New("CFS might not be deleted after this run")
		} else {
			return r.removeWithAttempts(ctx, attempt+1)
		}
	}

	return r.removeRole(ctx)
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

func (r *CloudFormationStack) doRemove() error { //nolint:gocyclo
	if r.parentID != nil {
		p, err := GetParentStack(r.svc, *r.parentID)
		if err != nil {
			return err
		}

		if p != nil {
			return liberrors.ErrHoldResource("waiting for parent stack")
		}
	}

	o, err := r.svc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: r.Name,
	})
	if err != nil {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			if awsErr.Code() == "ValidationFailed" && strings.HasSuffix(awsErr.Message(), " does not exist") {
				r.logger.Infof("CloudFormationStack stackName=%s no longer exists", *r.Name)
				return nil
			}
		}
		return err
	}
	stack := o.Stacks[0]

	if *stack.StackStatus == cloudformation.StackStatusDeleteComplete { //nolint:staticcheck
		// stack already deleted, no need to re-delete
		return nil
	} else if *stack.StackStatus == cloudformation.StackStatusDeleteInProgress {
		r.logger.Infof("CloudFormationStack stackName=%s delete in progress. Waiting", *r.Name)
		return r.svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: r.Name,
		})
	} else if *stack.StackStatus == cloudformation.StackStatusDeleteFailed {
		r.logger.Infof("CloudFormationStack stackName=%s delete failed (reason=%s). Attempting to retain and delete stack",
			*r.Name, ptr.ToString(stack.StackStatusReason))
		// This means the CFS has undetectable resources.
		// In order to move on with nuking, we retain them in the deletion.
		retainableResources, err := r.svc.ListStackResources(&cloudformation.ListStackResourcesInput{
			StackName: r.Name,
		})
		if err != nil {
			return err
		}

		retain := make([]*string, 0)

		for _, res := range retainableResources.StackResourceSummaries {
			if *res.ResourceStatus != cloudformation.ResourceStatusDeleteComplete {
				retain = append(retain, res.LogicalResourceId)
				r.logger.Infof("CloudFormationStack stackName=%s retaining resource %s (type=%s, status=%s, reason=%s)",
					*r.Name, ptr.ToString(res.LogicalResourceId), ptr.ToString(res.ResourceType),
					ptr.ToString(res.ResourceStatus), ptr.ToString(res.ResourceStatusReason))
			}
		}

		deleteInput := &cloudformation.DeleteStackInput{
			StackName:       r.Name,
			RetainResources: retain,
		}
		r.applyRoleOverride(deleteInput)

		if _, err = r.svc.DeleteStack(deleteInput); err != nil {
			return err
		}

		return r.svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: r.Name,
		})
	} else {
		if err := r.waitForStackToStabilize(*stack.StackStatus); err != nil {
			return err
		}

		deleteInput := &cloudformation.DeleteStackInput{
			StackName: r.Name,
		}
		r.applyRoleOverride(deleteInput)

		if _, err := r.svc.DeleteStack(deleteInput); err != nil {
			return err
		} else if err := r.svc.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
			StackName: r.Name,
		}); err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (r *CloudFormationStack) waitForStackToStabilize(currentStatus string) error {
	switch currentStatus {
	case cloudformation.StackStatusUpdateInProgress,
		cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateRollbackInProgress:
		r.logger.Infof("CloudFormationStack stackName=%s update in progress. Waiting to stabalize", *r.Name)

		return r.svc.WaitUntilStackUpdateComplete(&cloudformation.DescribeStacksInput{
			StackName: r.Name,
		})
	case cloudformation.StackStatusCreateInProgress,
		cloudformation.StackStatusRollbackInProgress:
		r.logger.Infof("CloudFormationStack stackName=%s create in progress. Waiting to stabalize", *r.Name)

		return r.svc.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
			StackName: r.Name,
		})
	default:
		return nil
	}
}

func (r *CloudFormationStack) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudFormationStack) String() string {
	return *r.Name
}
