package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gotidy/ptr"
	"go.uber.org/ratelimit"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchAlarmResource = "CloudWatchAlarm"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudWatchAlarmResource,
		Scope:    nuke.Account,
		Resource: &CloudWatchAlarm{},
		Lister:   &CloudWatchAlarmLister{},
	})
}

// ref - https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html

var CloudWatchAlarmDescribeRateLimit = ratelimit.New(8, ratelimit.Per(time.Second))
var CloudWatchAlarmDeleteRateLimit = ratelimit.New(3, ratelimit.Per(time.Second))
var CloudWatchAlarmListTagsRateLimit = ratelimit.New(5, ratelimit.Per(time.Second))

type CloudWatchAlarmLister struct{}

func (l *CloudWatchAlarmLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatch.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: []*string{
			ptr.String(cloudwatch.AlarmTypeCompositeAlarm),
			ptr.String(cloudwatch.AlarmTypeMetricAlarm),
		},
		MaxRecords: aws.Int64(100),
	}

	for {
		CloudWatchAlarmDescribeRateLimit.Take()

		output, err := svc.DescribeAlarms(params)
		if err != nil {
			return nil, err
		}

		for _, metricAlarm := range output.MetricAlarms {
			tags, err := GetAlarmTags(svc, metricAlarm.AlarmArn)
			if err != nil {
				return nil, err
			}
			resources = append(resources, &CloudWatchAlarm{
				svc:  svc,
				Name: metricAlarm.AlarmName,
				Type: ptr.String(cloudwatch.AlarmTypeMetricAlarm),
				Tags: tags,
			})
		}

		for _, compositeAlarm := range output.CompositeAlarms {
			tags, err := GetAlarmTags(svc, compositeAlarm.AlarmArn)
			if err != nil {
				return nil, err
			}
			resources = append(resources, &CloudWatchAlarm{
				svc:  svc,
				Name: compositeAlarm.AlarmName,
				Type: ptr.String(cloudwatch.AlarmTypeCompositeAlarm),
				Tags: tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func GetAlarmTags(svc *cloudwatch.CloudWatch, arn *string) ([]*cloudwatch.Tag, error) {
	CloudWatchAlarmListTagsRateLimit.Take()

	resp, err := svc.ListTagsForResource(&cloudwatch.ListTagsForResourceInput{ResourceARN: arn})
	if err != nil {
		return nil, err
	}

	return resp.Tags, nil
}

type CloudWatchAlarm struct {
	svc  *cloudwatch.CloudWatch
	Name *string
	Type *string
	Tags []*cloudwatch.Tag
}

func (r *CloudWatchAlarm) Remove(_ context.Context) error {
	CloudWatchAlarmDeleteRateLimit.Take()

	_, err := r.svc.DeleteAlarms(&cloudwatch.DeleteAlarmsInput{
		AlarmNames: []*string{r.Name},
	})

	return err
}

func (r *CloudWatchAlarm) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudWatchAlarm) String() string {
	return *r.Name
}
