package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                    //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchLogsDestinationResource = "CloudWatchLogsDestination"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudWatchLogsDestinationResource,
		Scope:    nuke.Account,
		Resource: &CloudWatchLogsDestination{},
		Lister:   &CloudWatchLogsDestinationLister{},
	})
}

type CloudWatchLogsDestinationLister struct{}

func (l *CloudWatchLogsDestinationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchlogs.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatchlogs.DescribeDestinationsInput{
		Limit: aws.Int64(50),
	}

	for {
		output, err := svc.DescribeDestinations(params)
		if err != nil {
			return nil, err
		}

		for _, destination := range output.Destinations {
			resources = append(resources, &CloudWatchLogsDestination{
				svc:             svc,
				destinationName: destination.DestinationName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CloudWatchLogsDestination struct {
	svc             *cloudwatchlogs.CloudWatchLogs
	destinationName *string
}

func (f *CloudWatchLogsDestination) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDestination(&cloudwatchlogs.DeleteDestinationInput{
		DestinationName: f.destinationName,
	})

	return err
}

func (f *CloudWatchLogsDestination) String() string {
	return *f.destinationName
}
