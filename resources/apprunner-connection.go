package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/apprunner"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppRunnerConnection struct {
	svc            *apprunner.AppRunner
	ConnectionArn  *string
	ConnectionName *string
}

const AppRunnerConnectionResource = "AppRunnerConnection"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppRunnerConnectionResource,
		Scope:  nuke.Account,
		Lister: &AppRunnerConnectionLister{},
	})
}

type AppRunnerConnectionLister struct{}

func (l *AppRunnerConnectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := apprunner.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &apprunner.ListConnectionsInput{}

	for {
		resp, err := svc.ListConnections(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.ConnectionSummaryList {
			resources = append(resources, &AppRunnerConnection{
				svc:            svc,
				ConnectionArn:  item.ConnectionArn,
				ConnectionName: item.ConnectionName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (f *AppRunnerConnection) Remove(_ context.Context) error {
	_, err := f.svc.DeleteConnection(&apprunner.DeleteConnectionInput{
		ConnectionArn: f.ConnectionArn,
	})

	return err
}

func (f *AppRunnerConnection) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ConnectionArn", f.ConnectionArn)
	properties.Set("ConnectionName", f.ConnectionName)
	return properties
}
