package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AmplifyAppResource = "AmplifyApp"

func init() {
	registry.Register(&registry.Registration{
		Name:     AmplifyAppResource,
		Scope:    nuke.Account,
		Resource: &AmplifyApp{},
		Lister:   &AmplifyAppLister{},
	})
}

type AmplifyAppLister struct{}

func (l *AmplifyAppLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := amplify.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &amplify.ListAppsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListApps(params)
		if err != nil {
			return nil, err
		}

		for _, item := range output.Apps {
			resources = append(resources, &AmplifyApp{
				svc:   svc,
				AppID: item.AppId,
				Name:  item.Name,
				Tags:  item.Tags,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type AmplifyApp struct {
	svc   *amplify.Amplify
	AppID *string
	Name  *string
	Tags  map[string]*string
}

func (r *AmplifyApp) Remove(_ context.Context) error {
	_, err := r.svc.DeleteApp(&amplify.DeleteAppInput{
		AppId: r.AppID,
	})

	return err
}

func (r *AmplifyApp) String() string {
	return *r.AppID
}

func (r *AmplifyApp) Properties() types.Properties {
	properties := types.NewProperties()
	for key, tag := range r.Tags {
		properties.SetTag(&key, tag)
	}
	properties.
		Set("AppID", r.AppID).
		Set("Name", r.Name)
	return properties
}
