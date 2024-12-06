package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MediaLiveChannelResource = "MediaLiveChannel"

func init() {
	registry.Register(&registry.Registration{
		Name:     MediaLiveChannelResource,
		Scope:    nuke.Account,
		Resource: &MediaLiveChannel{},
		Lister:   &MediaLiveChannelLister{},
	})
}

type MediaLiveChannelLister struct{}

func (l *MediaLiveChannelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := medialive.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &medialive.ListChannelsInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.ListChannels(params)
		if err != nil {
			return nil, err
		}

		for _, channel := range output.Channels {
			resources = append(resources, &MediaLiveChannel{
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

type MediaLiveChannel struct {
	svc *medialive.MediaLive
	ID  *string
}

func (f *MediaLiveChannel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteChannel(&medialive.DeleteChannelInput{
		ChannelId: f.ID,
	})

	return err
}

func (f *MediaLiveChannel) String() string {
	return *f.ID
}
