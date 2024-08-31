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

const CloudWatchAnomalyDetectorResource = "CloudWatchAnomalyDetector"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudWatchAnomalyDetectorResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchAnomalyDetectorLister{},
	})
}

type CloudWatchAnomalyDetectorLister struct{}

func (l *CloudWatchAnomalyDetectorLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatch.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatch.DescribeAnomalyDetectorsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		output, err := svc.DescribeAnomalyDetectors(params)
		if err != nil {
			return nil, err
		}

		for _, detector := range output.AnomalyDetectors {
			resources = append(resources, &CloudWatchAnomalyDetector{
				svc:        svc,
				detector:   detector.SingleMetricAnomalyDetector,
				MetricName: detector.SingleMetricAnomalyDetector.MetricName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CloudWatchAnomalyDetector struct {
	svc        *cloudwatch.CloudWatch
	detector   *cloudwatch.SingleMetricAnomalyDetector
	MetricName *string
}

func (r *CloudWatchAnomalyDetector) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAnomalyDetector(&cloudwatch.DeleteAnomalyDetectorInput{
		SingleMetricAnomalyDetector: r.detector,
	})

	return err
}

func (r *CloudWatchAnomalyDetector) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *CloudWatchAnomalyDetector) String() string {
	return *r.MetricName
}
