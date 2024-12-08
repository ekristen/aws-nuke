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

const AppConfigHostedConfigurationVersionResource = "AppConfigHostedConfigurationVersion"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppConfigHostedConfigurationVersionResource,
		Scope:    nuke.Account,
		Resource: &AppConfigHostedConfigurationVersion{},
		Lister:   &AppConfigHostedConfigurationVersionLister{},
	})
}

type AppConfigHostedConfigurationVersionLister struct{}

func (l *AppConfigHostedConfigurationVersionLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appconfig.New(opts.Session)
	resources := make([]resource.Resource, 0)

	profilerLister := &AppConfigConfigurationProfileLister{}
	configurationProfiles, err := profilerLister.List(ctx, o)
	if err != nil {
		return nil, err
	}
	for _, configurationProfileResource := range configurationProfiles {
		configurationProfile, ok := configurationProfileResource.(*AppConfigConfigurationProfile)
		if !ok {
			logrus.Errorf("Unable to cast AppConfigConfigurationProfile.")
			continue
		}
		params := &appconfig.ListHostedConfigurationVersionsInput{
			ApplicationId:          configurationProfile.applicationID,
			ConfigurationProfileId: configurationProfile.id,
			MaxResults:             aws.Int64(50),
		}
		err := svc.ListHostedConfigurationVersionsPages(params, func(page *appconfig.ListHostedConfigurationVersionsOutput, lastPage bool) bool {
			for _, item := range page.Items {
				resources = append(resources, &AppConfigHostedConfigurationVersion{
					svc:                    svc,
					applicationID:          configurationProfile.applicationID,
					configurationProfileID: configurationProfile.id,
					versionNumber:          item.VersionNumber,
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

type AppConfigHostedConfigurationVersion struct {
	svc                    *appconfig.AppConfig
	applicationID          *string
	configurationProfileID *string
	versionNumber          *int64
}

func (f *AppConfigHostedConfigurationVersion) Remove(_ context.Context) error {
	_, err := f.svc.DeleteHostedConfigurationVersion(&appconfig.DeleteHostedConfigurationVersionInput{
		ApplicationId:          f.applicationID,
		ConfigurationProfileId: f.configurationProfileID,
		VersionNumber:          f.versionNumber,
	})
	return err
}

func (f *AppConfigHostedConfigurationVersion) Properties() types.Properties {
	return types.NewProperties().
		Set("ApplicationID", f.applicationID).
		Set("ConfigurationProfileID", f.configurationProfileID).
		Set("VersionNumber", f.versionNumber)
}
