package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/apprunner"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppRunnerService struct {
	svc         *apprunner.AppRunner
	ServiceArn  *string
	ServiceId   *string
	ServiceName *string
}

const AppRunnerServiceResource = "AppRunnerService"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppRunnerServiceResource,
		Scope:  nuke.Account,
		Lister: &AppRunnerServiceLister{},
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
				ServiceArn:  item.ServiceArn,
				ServiceId:   item.ServiceId,
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

func (f *AppRunnerService) Remove(_ context.Context) error {
	_, err := f.svc.DeleteService(&apprunner.DeleteServiceInput{
		ServiceArn: f.ServiceArn,
	})

	return err
}

func (f *AppRunnerService) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ServiceArn", f.ServiceArn)
	properties.Set("ServiceId", f.ServiceId)
	properties.Set("ServiceName", f.ServiceName)
	return properties
}
