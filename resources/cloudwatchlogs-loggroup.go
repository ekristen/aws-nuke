package resources

import (
	"context"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"go.uber.org/ratelimit"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

// Note: this is global, it really should be per-region
var deleteRl = ratelimit.New(10)

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
	tagRl := ratelimit.New(5)
	streamRl := ratelimit.New(25)

	params := &cloudwatchlogs.DescribeLogGroupsInput{
		Limit: ptr.Int64(50),
	}

	for {
		groupRl.Take() // Wait for DescribeLogGroups rate limiter

		output, err := svc.DescribeLogGroups(params)
		if err != nil {
			return nil, err
		}

		for _, logGroup := range output.LogGroups {
			tagRl.Take() // Wait for ListTagsForResource rate limiter

			arn := strings.TrimSuffix(*logGroup.Arn, ":*")
			tagResp, err := svc.ListTagsForResource(
				&cloudwatchlogs.ListTagsForResourceInput{
					ResourceArn: &arn,
				})
			if err != nil {
				return nil, err
			}

			streamRl.Take() // Wait for DescribeLogStreams rate limiter

			// get last event ingestion time
			lsResp, err := svc.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
				LogGroupName: logGroup.LogGroupName,
				OrderBy:      ptr.String("LastEventTime"),
				Limit:        ptr.Int64(1),
				Descending:   ptr.Bool(true),
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

			var retentionInDays int64
			if logGroup.RetentionInDays != nil {
				retentionInDays = ptr.ToInt64(logGroup.RetentionInDays)
			}

			resources = append(resources, &CloudWatchLogsLogGroup{
				svc:             svc,
				Name:            logGroup.LogGroupName,
				CreatedTime:     logGroup.CreationTime,
				CreationTime:    ptr.Time(time.Unix(*logGroup.CreationTime/1000, 0).UTC()),
				LastEvent:       ptr.Time(lastEvent), // TODO(v4): convert to UTC
				RetentionInDays: retentionInDays,
				Tags:            tagResp.Tags,
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
	svc             *cloudwatchlogs.CloudWatchLogs
	Name            *string    `description:"The name of the log group" libnuke:"uniqueKey"`
	CreatedTime     *int64     `description:"The creation time of the log group in unix timestamp format"`
	CreationTime    *time.Time `description:"The creation time of the log group in RFC3339 format" libnuke:"uniqueKey"`
	LastEvent       *time.Time `description:"The last event time of the log group in RFC3339 format"`
	RetentionInDays int64      `description:"The number of days to retain log events in the log group"`
	Tags            map[string]*string
}

func (r *CloudWatchLogsLogGroup) Remove(_ context.Context) error {
	deleteRl.Take() // Wait for DeleteLogGroup rate limiter

	_, err := r.svc.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: r.Name,
	})

	return err
}

func (r *CloudWatchLogsLogGroup) String() string {
	return *r.Name
}

func (r *CloudWatchLogsLogGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r).
		Set("logGroupName", r.Name) // TODO(v4): remove this property
}
