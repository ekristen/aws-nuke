package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppConfigApplication struct {
	svc  *appconfig.AppConfig
	id   *string
	name *string
}

const AppConfigApplicationResource = "AppConfigApplication"

func init() {
	resource.Register(&resource.Registration{
		Name:   AppConfigApplicationResource,
		Scope:  nuke.Account,
		Lister: &AppConfigApplicationLister{},
		DependsOn: []string{
			AppConfigConfigurationProfileResource,
			AppConfigEnvironmentResource,
		},
	})
}

type AppConfigApplicationLister struct{}

func (l *AppConfigApplicationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appconfig.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &appconfig.ListApplicationsInput{
		MaxResults: aws.Int64(50),
	}
	err := svc.ListApplicationsPages(params, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		for _, item := range page.Items {
			resources = append(resources, &AppConfigApplication{
				svc:  svc,
				id:   item.Id,
				name: item.Name,
			})
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (f *AppConfigApplication) Remove(_ context.Context) error {
	_, err := f.svc.DeleteApplication(&appconfig.DeleteApplicationInput{
		ApplicationId: f.id,
	})
	return err
}

func (f *AppConfigApplication) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", f.id).
		Set("Name", f.name)
}
