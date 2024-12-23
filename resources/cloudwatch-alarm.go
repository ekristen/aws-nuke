package resources

import (
	"context"
	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

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
