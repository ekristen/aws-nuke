package resources

import (
	"context"

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
				svc:       svc,
				alarmName: metricAlarm.AlarmName,
				tags:      tags,
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
	svc       *cloudwatch.CloudWatch
	alarmName *string
	tags      []*cloudwatch.Tag
}

func (f *CloudWatchAlarm) Remove(_ context.Context) error {
	_, err := f.svc.DeleteAlarms(&cloudwatch.DeleteAlarmsInput{
		AlarmNames: []*string{f.alarmName},
	})

	return err
}

func (f *CloudWatchAlarm) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.alarmName)

	for _, tag := range f.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	return properties
}

func (f *CloudWatchAlarm) String() string {
	return *f.alarmName
}
