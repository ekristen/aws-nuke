package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/apprunner" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const AppRunnerServiceResource = "AppRunnerService"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppRunnerServiceResource,
		Scope:    nuke.Account,
		Resource: &AppRunnerService{},
		Lister:   &AppRunnerServiceLister{},
	})
}

type AppRunnerServiceLister struct{}

func (l *AppRunnerServiceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := apprunner.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &apprunner.ListServicesInput{}

	for {
		resp, err := svc.ListServices(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.ServiceSummaryList {
			resources = append(resources, &AppRunnerService{
				svc:         svc,
				ServiceARN:  item.ServiceArn,
				ServiceID:   item.ServiceId,
				ServiceName: item.ServiceName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type AppRunnerService struct {
	svc         *apprunner.AppRunner
	ServiceARN  *string
	ServiceID   *string
	ServiceName *string
}

func (f *AppRunnerService) Remove(_ context.Context) error {
	_, err := f.svc.DeleteService(&apprunner.DeleteServiceInput{
		ServiceArn: f.ServiceARN,
	})

	return err
}

func (f *AppRunnerService) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ServiceArn", f.ServiceARN)
	properties.Set("ServiceId", f.ServiceID)
	properties.Set("ServiceName", f.ServiceName)
	return properties
}
