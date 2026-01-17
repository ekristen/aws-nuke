package resources

import (
	"context"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
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
		Settings: []string{
			"DisableDeletionProtection",
		},
		DependsOn: []string{
			EC2VPCResource,         // Reason: flow logs, if log group is cleaned before vpc, vpc can write more flow logs
			LambdaFunctionResource, // Reason: Lambda functions can recreate log groups due to invocations, automatic container provisioning, etc.
		},
	})
}

type CloudWatchLogsLogGroupLister struct{}

func (l *CloudWatchLogsLogGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchlogs.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	// Note: these can be modified by the customer and account, and we could query them but for now we hard code
	// them to the bottom, because really it's per-second, and we should be fine querying at this rate for clearing
	// Ref: https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/cloudwatch_limits_cwl.html
	groupRl := ratelimit.New(10)
	tagRl := ratelimit.New(5)
	streamRl := ratelimit.New(25)

	params := &cloudwatchlogs.DescribeLogGroupsInput{
		Limit: aws.Int32(50),
	}

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(svc, params)

	for paginator.HasMorePages() {
		groupRl.Take() // Wait for DescribeLogGroups rate limiter

		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range output.LogGroups {
			logGroup := &output.LogGroups[i]
			tagRl.Take() // Wait for ListTagsForResource rate limiter

			arn := strings.TrimSuffix(*logGroup.Arn, ":*")
			tagResp, err := svc.ListTagsForResource(ctx,
				&cloudwatchlogs.ListTagsForResourceInput{
					ResourceArn: &arn,
				})
			if err != nil {
				logrus.WithError(err).
					WithField("arn", arn).
					Warn("unable to list tags for log group, skipping to avoid incorrect filtering")
				continue
			}

			streamRl.Take() // Wait for DescribeLogStreams rate limiter

			// get last event ingestion time
			lsResp, err := svc.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
				LogGroupName: logGroup.LogGroupName,
				OrderBy:      "LastEventTime",
				Limit:        aws.Int32(1),
				Descending:   aws.Bool(true),
			})
			if err != nil {
				logrus.WithError(err).
					WithField("arn", arn).
					Warn("unable to describe log streams for log group, skipping to avoid incorrect filtering")
				continue
			}

			var lastEvent time.Time
			if len(lsResp.LogStreams) > 0 && lsResp.LogStreams[0].LastIngestionTime != nil {
				lastEvent = time.Unix(*lsResp.LogStreams[0].LastIngestionTime/1000, 0)
			} else {
				lastEvent = time.Unix(*logGroup.CreationTime/1000, 0)
			}

			var retentionInDays int64
			if logGroup.RetentionInDays != nil {
				retentionInDays = int64(ptr.ToInt32(logGroup.RetentionInDays))
			}

			resources = append(resources, &CloudWatchLogsLogGroup{
				svc:             svc,
				Name:            logGroup.LogGroupName,
				CreatedTime:     logGroup.CreationTime,
				CreationTime:    ptr.Time(time.Unix(*logGroup.CreationTime/1000, 0).UTC()),
				LastEvent:       ptr.Time(lastEvent), // TODO(v4): convert to UTC
				RetentionInDays: retentionInDays,
				Tags:            tagResp.Tags,
				protection:      logGroup.DeletionProtectionEnabled,
			})
		}
	}

	return resources, nil
}

type CloudWatchLogsLogGroup struct {
	svc             *cloudwatchlogs.Client
	Name            *string    `description:"The name of the log group" libnuke:"uniqueKey"`
	CreatedTime     *int64     `description:"The creation time of the log group in unix timestamp format"`
	CreationTime    *time.Time `description:"The creation time of the log group in RFC3339 format" libnuke:"uniqueKey"`
	LastEvent       *time.Time `description:"The last event time of the log group in RFC3339 format"`
	RetentionInDays int64      `description:"The number of days to retain log events in the log group"`
	Tags            map[string]string
	settings        *libsettings.Setting
	protection      *bool
}

func (r *CloudWatchLogsLogGroup) Remove(ctx context.Context) error {
	if ptr.ToBool(r.protection) && r.settings.GetBool("DisableDeletionProtection") {
		_, err := r.svc.PutLogGroupDeletionProtection(ctx, &cloudwatchlogs.PutLogGroupDeletionProtectionInput{
			LogGroupIdentifier:        r.Name,
			DeletionProtectionEnabled: aws.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	deleteRl.Take() // Wait for DeleteLogGroup rate limiter

	_, err := r.svc.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
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

func (r *CloudWatchLogsLogGroup) Settings(setting *libsettings.Setting) {
	r.settings = setting
}
