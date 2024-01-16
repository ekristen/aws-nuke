package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTStreamResource = "IoTStream"

func init() {
	resource.Register(resource.Registration{
		Name:   IoTStreamResource,
		Scope:  nuke.Account,
		Lister: &IoTStreamLister{},
	})
}

type IoTStreamLister struct{}

func (l *IoTStreamLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListStreamsInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListStreams(params)
		if err != nil {
			return nil, err
		}

		for _, stream := range output.Streams {
			resources = append(resources, &IoTStream{
				svc: svc,
				ID:  stream.StreamId,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type IoTStream struct {
	svc *iot.IoT
	ID  *string
}

func (f *IoTStream) Remove(_ context.Context) error {
	_, err := f.svc.DeleteStream(&iot.DeleteStreamInput{
		StreamId: f.ID,
	})

	return err
}

func (f *IoTStream) String() string {
	return *f.ID
}
