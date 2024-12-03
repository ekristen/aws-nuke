package resources

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"go.uber.org/ratelimit"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchLogsLogGroupResource = "CloudWatchLogsLogGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudWatchLogsLogGroupResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchLogsLogGroupLister{},
		DependsOn: []string{
			EC2VPCResource, // Reason: flow logs, if log group is cleaned before vpc, vpc can write more flow logs
		},
	})
}

type CloudWatchLogsLogGroupLister struct{}

func (l *CloudWatchLogsLogGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchlogs.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Note: these can be modified by the customer and account, and we could query them but for now we hard code
	// them to the bottom, because really it's per-second, and we should be fine querying at this rate for clearing
	// Ref: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/cloudwatch_limits_cwl.html
	groupRl := ratelimit.New(10)

	params := &cloudwatchlogs.DescribeLogGroupsInput{
		Limit: aws.Int64(50),
	}

	pageLimit := 20 // limit to 50*20 = 1000 log groups

	for {
		groupRl.Take() // Wait for DescribeLogGroup rate limiter

		output, err := svc.DescribeLogGroups(params)
		if err != nil {
			return nil, err
		}

		for _, logGroup := range output.LogGroups {
			resources = append(resources, &CloudWatchLogsLogGroup{
				svc:      svc,
				logGroup: logGroup,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken

		if pageLimit == 0 {
			break
		}

		pageLimit--
	}

	return resources, nil
}

type CloudWatchLogsLogGroup struct {
	svc       *cloudwatchlogs.CloudWatchLogs
	logGroup  *cloudwatchlogs.LogGroup
	lastEvent string
	tags      map[string]*string
}

func (f *CloudWatchLogsLogGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: f.logGroup.LogGroupName,
	})

	return err
}

func (f *CloudWatchLogsLogGroup) String() string {
	return *f.logGroup.LogGroupName
}

func (f *CloudWatchLogsLogGroup) Properties() types.Properties {
	properties := types.NewProperties().
		Set("logGroupName", f.logGroup.LogGroupName).
		Set("CreatedTime", f.logGroup.CreationTime).
		Set("LastEvent", f.lastEvent)

	for k, v := range f.tags {
		properties.SetTag(&k, v)
	}
	return properties
}
