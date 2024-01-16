package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediatailor"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaTailorConfigurationResource = "MediaTailorConfiguration"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaTailorConfigurationResource,
		Scope:  nuke.Account,
		Lister: &MediaTailorConfigurationLister{},
	})
}

type MediaTailorConfigurationLister struct{}

func (l *MediaTailorConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediatailor.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &mediatailor.ListPlaybackConfigurationsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.ListPlaybackConfigurations(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.Items {
			resources = append(resources, &MediaTailorConfiguration{
				svc:  svc,
				name: item.Name,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type MediaTailorConfiguration struct {
	svc  *mediatailor.MediaTailor
	name *string
}

func (f *MediaTailorConfiguration) Remove(_ context.Context) error {
	_, err := f.svc.DeletePlaybackConfiguration(&mediatailor.DeletePlaybackConfigurationInput{
		Name: f.name,
	})

	return err
}

func (f *MediaTailorConfiguration) String() string {
	return *f.name
}
