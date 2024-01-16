package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const FirehoseDeliveryStreamResource = "FirehoseDeliveryStream"

func init() {
	resource.Register(resource.Registration{
		Name:   FirehoseDeliveryStreamResource,
		Scope:  nuke.Account,
		Lister: &FirehoseDeliveryStreamLister{},
	})
}

type FirehoseDeliveryStreamLister struct{}

func (l *FirehoseDeliveryStreamLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := firehose.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var lastDeliveryStreamName *string

	params := &firehose.ListDeliveryStreamsInput{
		Limit: aws.Int64(25),
	}

	for {
		output, err := svc.ListDeliveryStreams(params)
		if err != nil {
			return nil, err
		}

		for _, deliveryStreamName := range output.DeliveryStreamNames {
			resources = append(resources, &FirehoseDeliveryStream{
				svc:                svc,
				deliveryStreamName: deliveryStreamName,
			})
			lastDeliveryStreamName = deliveryStreamName
		}

		if *output.HasMoreDeliveryStreams == false {
			break
		}

		params.ExclusiveStartDeliveryStreamName = lastDeliveryStreamName
	}

	return resources, nil
}

type FirehoseDeliveryStream struct {
	svc                *firehose.Firehose
	deliveryStreamName *string
}

func (f *FirehoseDeliveryStream) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDeliveryStream(&firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: f.deliveryStreamName,
	})

	return err
}

func (f *FirehoseDeliveryStream) String() string {
	return *f.deliveryStreamName
}
