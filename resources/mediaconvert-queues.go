package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaConvertQueueResource = "MediaConvertQueue"

func init() {
	resource.Register(&resource.Registration{
		Name:   MediaConvertQueueResource,
		Scope:  nuke.Account,
		Lister: &MediaConvertQueueLister{},
	})
}

type MediaConvertQueueLister struct{}

func (l *MediaConvertQueueLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediaconvert.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var mediaEndpoint *string

	output, err := svc.DescribeEndpoints(&mediaconvert.DescribeEndpointsInput{})
	if err != nil {
		return nil, err
	}

	for _, endpoint := range output.Endpoints {
		mediaEndpoint = endpoint.Url
	}

	// Update svc to use custom media endpoint
	svc.Endpoint = *mediaEndpoint

	params := &mediaconvert.ListQueuesInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.ListQueues(params)
		if err != nil {
			return nil, err
		}

		for _, queue := range output.Queues {
			resources = append(resources, &MediaConvertQueue{
				svc:  svc,
				name: queue.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaConvertQueue struct {
	svc  *mediaconvert.MediaConvert
	name *string
}

func (f *MediaConvertQueue) Remove(_ context.Context) error {
	_, err := f.svc.DeleteQueue(&mediaconvert.DeleteQueueInput{
		Name: f.name,
	})

	return err
}

func (f *MediaConvertQueue) String() string {
	return *f.name
}

func (f *MediaConvertQueue) Filter() error {
	if strings.Contains(*f.name, "Default") {
		return fmt.Errorf("cannot delete default queue")
	}
	return nil
}
