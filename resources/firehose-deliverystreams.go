package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const FirehoseDeliveryStreamResource = "FirehoseDeliveryStream"

func init() {
	registry.Register(&registry.Registration{
		Name:     FirehoseDeliveryStreamResource,
		Scope:    nuke.Account,
		Resource: &FirehoseDeliveryStream{},
		Lister:   &FirehoseDeliveryStreamLister{},
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
			var tags []*firehose.Tag
			tagResp, err := svc.ListTagsForDeliveryStream(
				&firehose.ListTagsForDeliveryStreamInput{
					DeliveryStreamName: deliveryStreamName,
				})
			if err != nil {
				logrus.WithError(err).Error("unable to list tags for delivery stream")
			}
			if tagResp != nil {
				tags = tagResp.Tags
			}

			resources = append(resources, &FirehoseDeliveryStream{
				svc:                svc,
				deliveryStreamName: deliveryStreamName,
				tags:               tags,
			})
			lastDeliveryStreamName = deliveryStreamName
		}

		if !ptr.ToBool(output.HasMoreDeliveryStreams) {
			break
		}

		params.ExclusiveStartDeliveryStreamName = lastDeliveryStreamName
	}

	return resources, nil
}

type FirehoseDeliveryStream struct {
	svc                *firehose.Firehose
	deliveryStreamName *string
	tags               []*firehose.Tag
}

func (f *FirehoseDeliveryStream) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDeliveryStream(&firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: f.deliveryStreamName,
	})

	return err
}

func (f *FirehoseDeliveryStream) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range f.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", f.deliveryStreamName)
	return properties
}

func (f *FirehoseDeliveryStream) String() string {
	return *f.deliveryStreamName
}
