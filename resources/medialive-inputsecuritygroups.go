package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/medialive"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MediaLiveInputSecurityGroupResource = "MediaLiveInputSecurityGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   MediaLiveInputSecurityGroupResource,
		Scope:  nuke.Account,
		Lister: &MediaLiveInputSecurityGroupLister{},
	})
}

type MediaLiveInputSecurityGroupLister struct{}

func (l *MediaLiveInputSecurityGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := medialive.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &medialive.ListInputSecurityGroupsInput{
		MaxResults: aws.Int64(20),
	}

	for {
		output, err := svc.ListInputSecurityGroups(params)
		if err != nil {
			return nil, err
		}

		for _, inputSecurityGroup := range output.InputSecurityGroups {
			resources = append(resources, &MediaLiveInputSecurityGroup{
				svc: svc,
				ID:  inputSecurityGroup.Id,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MediaLiveInputSecurityGroup struct {
	svc *medialive.MediaLive
	ID  *string
}

func (f *MediaLiveInputSecurityGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteInputSecurityGroup(&medialive.DeleteInputSecurityGroupInput{
		InputSecurityGroupId: f.ID,
	})

	return err
}

func (f *MediaLiveInputSecurityGroup) String() string {
	return *f.ID
}
