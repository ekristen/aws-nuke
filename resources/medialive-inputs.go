package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaLiveInputResource = "MediaLiveInput"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaLiveInputResource,
		Scope:  nuke.Account,
		Lister: &MediaLiveInputLister{},
	})
}

type MediaLiveInputLister struct{}

func (l *MediaLiveInputLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := medialive.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &medialive.ListInputsInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.ListInputs(params)
		if err != nil {
			return nil, err
		}

		for _, input := range output.Inputs {
			resources = append(resources, &MediaLiveInput{
				svc: svc,
				ID:  input.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaLiveInput struct {
	svc *medialive.MediaLive
	ID  *string
}

func (f *MediaLiveInput) Remove(_ context.Context) error {
	_, err := f.svc.DeleteInput(&medialive.DeleteInputInput{
		InputId: f.ID,
	})

	return err
}

func (f *MediaLiveInput) String() string {
	return *f.ID
}
