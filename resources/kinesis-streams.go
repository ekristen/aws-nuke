package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const KinesisStreamResource = "KinesisStream"

func init() {
	registry.Register(&registry.Registration{
		Name:     KinesisStreamResource,
		Scope:    nuke.Account,
		Resource: &KinesisStream{},
		Lister:   &KinesisStreamLister{},
	})
}

type KinesisStreamAPI interface {
	ListStreams(ctx context.Context, params *kinesis.ListStreamsInput,
		optFns ...func(*kinesis.Options)) (*kinesis.ListStreamsOutput, error)
	ListTagsForStream(ctx context.Context, params *kinesis.ListTagsForStreamInput,
		optFns ...func(*kinesis.Options)) (*kinesis.ListTagsForStreamOutput, error)
	DeleteStream(ctx context.Context, params *kinesis.DeleteStreamInput,
		optFns ...func(*kinesis.Options)) (*kinesis.DeleteStreamOutput, error)
}

type KinesisStreamLister struct {
	svc KinesisStreamAPI
}

func (l *KinesisStreamLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	if l.svc == nil {
		l.svc = kinesis.NewFromConfig(*opts.Config)
	}

	resources := make([]resource.Resource, 0)
	var lastStreamName *string

	params := &kinesis.ListStreamsInput{
		Limit: aws.Int32(25),
	}

	for {
		output, err := l.svc.ListStreams(ctx, params)
		if err != nil {
			return nil, err
		}

		for _, streamName := range output.StreamNames {
			name := streamName

			tags := make(map[string]string)
			tagsResp, err := l.svc.ListTagsForStream(ctx, &kinesis.ListTagsForStreamInput{
				StreamName: &name,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for kinesis stream: %s", name)
			} else {
				for _, tag := range tagsResp.Tags {
					tags[*tag.Key] = *tag.Value
				}
			}

			resources = append(resources, &KinesisStream{
				svc:  l.svc,
				Name: &name,
				Tags: tags,
			})
			lastStreamName = &name
		}

		if output.HasMoreStreams == nil || !*output.HasMoreStreams {
			break
		}

		params.ExclusiveStartStreamName = lastStreamName
	}

	return resources, nil
}

type KinesisStream struct {
	svc  KinesisStreamAPI
	Name *string           `description:"The name of the Kinesis stream"`
	Tags map[string]string `description:"The tags associated with the Kinesis stream"`
}

func (f *KinesisStream) Remove(ctx context.Context) error {
	_, err := f.svc.DeleteStream(ctx, &kinesis.DeleteStreamInput{
		StreamName: f.Name,
	})

	return err
}

func (f *KinesisStream) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *KinesisStream) String() string {
	return *f.Name
}
