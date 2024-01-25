package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppConfigEnvironment struct {
	svc           *appconfig.AppConfig
	applicationId *string
	id            *string
	name          *string
}

const AppConfigEnvironmentResource = "AppConfigEnvironment"

func init() {
	resource.Register(&resource.Registration{
		Name:   AppConfigEnvironmentResource,
		Scope:  nuke.Account,
		Lister: &AppConfigEnvironmentLister{},
	})
}

type AppConfigEnvironmentLister struct{}

func (l *AppConfigEnvironmentLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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
		params := &appconfig.ListEnvironmentsInput{
			ApplicationId: application.id,
			MaxResults:    aws.Int64(50),
		}
		err := svc.ListEnvironmentsPages(params, func(page *appconfig.ListEnvironmentsOutput, lastPage bool) bool {
			for _, item := range page.Items {
				resources = append(resources, &AppConfigEnvironment{
					svc:           svc,
					applicationId: application.id,
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

func (f *AppConfigEnvironment) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEnvironment(&appconfig.DeleteEnvironmentInput{
		ApplicationId: f.applicationId,
		EnvironmentId: f.id,
	})
	return err
}

func (f *AppConfigEnvironment) Properties() types.Properties {
	return types.NewProperties().
		Set("ApplicationID", f.applicationId).
		Set("ID", f.id).
		Set("Name", f.name)
}
