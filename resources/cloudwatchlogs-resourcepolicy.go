package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                    //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchLogsResourcePolicyResource = "CloudWatchLogsResourcePolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudWatchLogsResourcePolicyResource,
		Scope:    nuke.Account,
		Resource: &CloudWatchLogsResourcePolicy{},
		Lister:   &CloudWatchLogsResourcePolicyLister{},
	})
}

type CloudWatchLogsResourcePolicyLister struct{}

func (l *CloudWatchLogsResourcePolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchlogs.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatchlogs.DescribeResourcePoliciesInput{
		Limit: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeResourcePolicies(params)
		if err != nil {
			return nil, err
		}

		for _, destination := range output.ResourcePolicies {
			resources = append(resources, &CloudWatchLogsResourcePolicy{
				svc:        svc,
				policyName: destination.PolicyName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CloudWatchLogsResourcePolicy struct {
	svc        *cloudwatchlogs.CloudWatchLogs
	policyName *string
}

func (p *CloudWatchLogsResourcePolicy) Remove(_ context.Context) error {
	_, err := p.svc.DeleteResourcePolicy(&cloudwatchlogs.DeleteResourcePolicyInput{
		PolicyName: p.policyName,
	})

	return err
}

func (p *CloudWatchLogsResourcePolicy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", p.policyName)

	return properties
}

func (p *CloudWatchLogsResourcePolicy) String() string {
	return *p.policyName
}
