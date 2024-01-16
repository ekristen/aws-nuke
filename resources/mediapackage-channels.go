package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaPackageChannelResource = "MediaPackageChannel"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaPackageChannelResource,
		Scope:  nuke.Account,
		Lister: &MediaPackageChannelLister{},
	})
}

type MediaPackageChannelLister struct{}

func (l *MediaPackageChannelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediapackage.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mediapackage.ListChannelsInput{
		MaxResults: aws.Int64(50),
	}

	for {
		output, err := svc.ListChannels(params)
		if err != nil {
			return nil, err
		}

		for _, channel := range output.Channels {
			resources = append(resources, &MediaPackageChannel{
				svc: svc,
				ID:  channel.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaPackageChannel struct {
	svc *mediapackage.MediaPackage
	ID  *string
}

func (f *MediaPackageChannel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteChannel(&mediapackage.DeleteChannelInput{
		Id: f.ID,
	})

	return err
}

func (f *MediaPackageChannel) String() string {
	return *f.ID
}
