package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFormationStackSetResource = "CloudFormationStackSet"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFormationStackSetResource,
		Scope:    nuke.Account,
		Resource: &CloudFormationStackSet{},
		Lister:   &CloudFormationStackSetLister{},
	})
}

type CloudFormationStackSetLister struct{}

func (l *CloudFormationStackSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudformation.New(opts.Session)

	params := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListStackSets(params)
		if err != nil {
			return nil, err
		}
		for _, stackSetSummary := range resp.Summaries {
			resources = append(resources, &CloudFormationStackSet{
				svc:             svc,
				stackSetSummary: stackSetSummary,
				sleepDuration:   10 * time.Second,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type CloudFormationStackSet struct {
	svc             cloudformationiface.CloudFormationAPI
	stackSetSummary *cloudformation.StackSetSummary
	sleepDuration   time.Duration
}

func (cfs *CloudFormationStackSet) findStackInstances() (map[string][]string, error) {
	accounts := make(map[string][]string)

	input := &cloudformation.ListStackInstancesInput{
		StackSetName: cfs.stackSetSummary.StackSetName,
	}

	for {
		resp, err := cfs.svc.ListStackInstances(input)
		if err != nil {
			return nil, err
		}
		for _, stackInstanceSummary := range resp.Summaries {
			if regions, ok := accounts[*stackInstanceSummary.Account]; !ok {
				accounts[*stackInstanceSummary.Account] = []string{*stackInstanceSummary.Region}
			} else {
				accounts[*stackInstanceSummary.Account] = append(regions, *stackInstanceSummary.Region)
			}
		}

		if resp.NextToken == nil {
			break
		}

		input.NextToken = resp.NextToken
	}

	return accounts, nil
}

func (cfs *CloudFormationStackSet) waitForStackSetOperation(operationID string) error {
	for {
		result, err := cfs.svc.DescribeStackSetOperation(&cloudformation.DescribeStackSetOperationInput{
			StackSetName: cfs.stackSetSummary.StackSetName,
			OperationId:  &operationID,
		})
		if err != nil {
			return err
		}
		logrus.Infof("Got stackInstance operation status on stackSet=%s operationID=%s status=%s",
			*cfs.stackSetSummary.StackSetName, operationID, *result.StackSetOperation.Status)

		if *result.StackSetOperation.Status == cloudformation.StackSetOperationResultStatusSucceeded { //nolint:staticcheck
			return nil
		} else if *result.StackSetOperation.Status == cloudformation.StackSetOperationResultStatusFailed ||
			*result.StackSetOperation.Status == cloudformation.StackSetOperationResultStatusCancelled {
			return fmt.Errorf("unable to delete stackSet=%s operationID=%s status=%s",
				*cfs.stackSetSummary.StackSetName, operationID, *result.StackSetOperation.Status)
		} else {
			logrus.Infof("Waiting on stackSet=%s operationID=%s status=%s",
				*cfs.stackSetSummary.StackSetName, operationID, *result.StackSetOperation.Status)

			time.Sleep(cfs.sleepDuration)
		}
	}
}

func (cfs *CloudFormationStackSet) deleteStackInstances(accountID string, regions []string) error {
	logrus.Infof("Deleting stack instance accountID=%s regions=%s", accountID, strings.Join(regions, ","))
	regionsInput := make([]*string, len(regions))
	for i, region := range regions {
		regionsInput[i] = aws.String(region)
		fmt.Printf("region=%s i=%d\n", region, i)
	}
	result, err := cfs.svc.DeleteStackInstances(&cloudformation.DeleteStackInstancesInput{
		StackSetName: cfs.stackSetSummary.StackSetName,
		Accounts:     []*string{&accountID},
		Regions:      regionsInput,
		// this will remove the stack set instance from the stackset, but will leave the stack
		// in the account/region it was deployed to
		RetainStacks: aws.Bool(true),
	})

	fmt.Printf("got result=%v err=%v\n", result, err)

	if result == nil {
		return fmt.Errorf("got null result")
	}
	if err != nil {
		return err
	}

	return cfs.waitForStackSetOperation(*result.OperationId)
}

func (cfs *CloudFormationStackSet) Remove(_ context.Context) error {
	accounts, err := cfs.findStackInstances()
	if err != nil {
		return err
	}
	for accountID, regions := range accounts {
		err := cfs.deleteStackInstances(accountID, regions)
		if err != nil {
			return err
		}
	}
	_, err = cfs.svc.DeleteStackSet(&cloudformation.DeleteStackSetInput{
		StackSetName: cfs.stackSetSummary.StackSetName,
	})
	return err
}

func (cfs *CloudFormationStackSet) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", cfs.stackSetSummary.StackSetName)
	properties.Set("StackSetId", cfs.stackSetSummary.StackSetId)

	return properties
}

func (cfs *CloudFormationStackSet) String() string {
	return *cfs.stackSetSummary.StackSetName
}
