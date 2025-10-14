package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                  //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/kinesisvideo" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const KinesisVideoProjectResource = "KinesisVideoProject"

func init() {
	registry.Register(&registry.Registration{
		Name:     KinesisVideoProjectResource,
		Scope:    nuke.Account,
		Resource: &KinesisVideoProject{},
		Lister:   &KinesisVideoProjectLister{},
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
