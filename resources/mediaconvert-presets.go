package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaConvertPresetResource = "MediaConvertPreset"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaConvertPresetResource,
		Scope:  nuke.Account,
		Lister: &MediaConvertPresetLister{},
	})
}

type MediaConvertPresetLister struct{}

func (l *MediaConvertPresetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := mediaconvert.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var mediaEndpoint *string

	output, err := svc.DescribeEndpoints(&mediaconvert.DescribeEndpointsInput{})
	if err != nil {
		return nil, err
	}

	for _, endpoint := range output.Endpoints {
		mediaEndpoint = endpoint.Url
	}

	// Update svc to use custom media endpoint
	svc.Endpoint = *mediaEndpoint

	params := &mediaconvert.ListPresetsInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.ListPresets(params)
		if err != nil {
			return nil, err
		}

		for _, preset := range output.Presets {
			resources = append(resources, &MediaConvertPreset{
				svc:  svc,
				name: preset.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaConvertPreset struct {
	svc  *mediaconvert.MediaConvert
	name *string
}

func (f *MediaConvertPreset) Remove(_ context.Context) error {
	_, err := f.svc.DeletePreset(&mediaconvert.DeletePresetInput{
		Name: f.name,
	})

	return err
}

func (f *MediaConvertPreset) String() string {
	return *f.name
}
