package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const KinesisVideoProjectResource = "KinesisVideoProject"

func init() {
	resource.Register(&resource.Registration{
		Name:   KinesisVideoProjectResource,
		Scope:  nuke.Account,
		Lister: &KinesisVideoProjectLister{},
	})
}

type KinesisVideoProjectLister struct{}

func (l *KinesisVideoProjectLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := kinesisvideo.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &kinesisvideo.ListStreamsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListStreams(params)
		if err != nil {
			return nil, err
		}

		for _, streamInfo := range output.StreamInfoList {
			resources = append(resources, &KinesisVideoProject{
				svc:       svc,
				streamARN: streamInfo.StreamARN,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type KinesisVideoProject struct {
	svc       *kinesisvideo.KinesisVideo
	streamARN *string
}

func (f *KinesisVideoProject) Remove(_ context.Context) error {
	_, err := f.svc.DeleteStream(&kinesisvideo.DeleteStreamInput{
		StreamARN: f.streamARN,
	})

	return err
}

func (f *KinesisVideoProject) String() string {
	return *f.streamARN
}
