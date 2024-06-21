package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type AppConfigConfigurationProfile struct {
	svc           *appconfig.AppConfig
	applicationID *string
	id            *string
	name          *string
}

const AppConfigConfigurationProfileResource = "AppConfigConfigurationProfile"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppConfigConfigurationProfileResource,
		Scope:  nuke.Account,
		Lister: &AppConfigConfigurationProfileLister{},
		DependsOn: []string{
			AppConfigHostedConfigurationVersionResource,
		},
	})
}

type AppConfigConfigurationProfileLister struct{}

func (l *AppConfigConfigurationProfileLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appconfig.New(opts.Session)
	resources := make([]resource.Resource, 0)

	applicationLister := &AppConfigApplicationLister{}

	applications, err := applicationLister.List(ctx, o)
	if err != nil {
		return nil, err
	}
	for _, applicationResource := range applications {
		application, ok := applicationResource.(*AppConfigApplication)
		if !ok {
			logrus.Errorf("Unable to cast AppConfigApplication.")
			continue
		}
		params := &appconfig.ListConfigurationProfilesInput{
			ApplicationId: application.id,
			MaxResults:    aws.Int64(50),
		}
		err := svc.ListConfigurationProfilesPages(params, func(page *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
			for _, item := range page.Items {
				resources = append(resources, &AppConfigConfigurationProfile{
					svc:           svc,
					applicationID: application.id,
					id:            item.Id,
					name:          item.Name,
				})
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	}
	return resources, nil
}

func (f *AppConfigConfigurationProfile) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConfigurationProfile(&appconfig.DeleteConfigurationProfileInput{
		ApplicationId:          f.applicationID,
		ConfigurationProfileId: f.id,
	})
	return err
}

func (f *AppConfigConfigurationProfile) Properties() types.Properties {
	return types.NewProperties().
		Set("ApplicationID", f.applicationID).
		Set("ID", f.id).
		Set("Name", f.name)
}
