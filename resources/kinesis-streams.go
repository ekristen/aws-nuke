package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const KinesisStreamResource = "KinesisStream"

func init() {
	registry.Register(&registry.Registration{
		Name:   KinesisStreamResource,
		Scope:  nuke.Account,
		Lister: &KinesisStreamLister{},
	})
}

type KinesisStreamLister struct{}

func (l *KinesisStreamLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kinesis.New(opts.Session)

	resources := make([]resource.Resource, 0)
	var lastStreamName *string

	params := &kinesis.ListStreamsInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListStreams(params)
		if err != nil {
			return nil, err
		}

		for _, streamName := range output.StreamNames {
			resources = append(resources, &KinesisStream{
				svc:        svc,
				streamName: streamName,
			})
			lastStreamName = streamName
		}

		if !aws.BoolValue(output.HasMoreStreams) {
			break
		}

		params.ExclusiveStartStreamName = lastStreamName
	}

	return resources, nil
}

type KinesisStream struct {
	svc        *kinesis.Kinesis
	streamName *string
}

func (f *KinesisStream) Remove(_ context.Context) error {
	_, err := f.svc.DeleteStream(&kinesis.DeleteStreamInput{
		StreamName: f.streamName,
	})

	return err
}

func (f *KinesisStream) String() string {
	return *f.streamName
}
