package resources

import (
	"context"
	"strings"
	"time"

	"go.uber.org/ratelimit"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchLogsLogGroupResource = "CloudWatchLogsLogGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudWatchLogsLogGroupResource,
		Scope:    nuke.Account,
		Resource: &CloudWatchLogsLogGroup{},
		Lister:   &CloudWatchLogsLogGroupLister{},
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
	streamRl := ratelimit.New(15)

	params := &cloudwatchlogs.DescribeLogGroupsInput{
		Limit: aws.Int64(50),
	}

	for {
		groupRl.Take() // Wait for DescribeLogGroup rate limiter

		output, err := svc.DescribeLogGroups(params)
		if err != nil {
			return nil, err
		}

		for _, logGroup := range output.LogGroups {
			streamRl.Take() // Wait for DescribeLogStream rate limiter

			arn := strings.TrimSuffix(*logGroup.Arn, ":*")
			tagResp, err := svc.ListTagsForResource(
				&cloudwatchlogs.ListTagsForResourceInput{
					ResourceArn: &arn,
				})
			if err != nil {
				return nil, err
			}

			// get last event ingestion time
			lsResp, err := svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
				LogGroupName: logGroup.LogGroupName,
				OrderBy:      aws.String("LastEventTime"),
				Limit:        aws.Int64(1),
				Descending:   aws.Bool(true),
			})
			if err != nil {
				return nil, err
			}

			var lastEvent time.Time
			if len(lsResp.LogStreams) > 0 && lsResp.LogStreams[0].LastIngestionTime != nil {
				lastEvent = time.Unix(*lsResp.LogStreams[0].LastIngestionTime/1000, 0)
			} else {
				lastEvent = time.Unix(*logGroup.CreationTime/1000, 0)
			}

			}

			resources = append(resources, &CloudWatchLogsLogGroup{
				svc:       svc,
				logGroup:  logGroup,
				lastEvent: lastEvent.Format(time.RFC3339),
				tags:      tagResp.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
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
		Set("RetentionInDays", f.logGroup.RetentionInDays)


	for k, v := range f.tags {
		properties.SetTag(&k, v)
	}
	return properties
}

