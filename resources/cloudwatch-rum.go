package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudwatchrum"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudWatchRUMAppResource = "CloudWatchRUMApp"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudWatchRUMAppResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchRUMAppLister{},
	})
}

type CloudWatchRUMAppLister struct{}

func (l *CloudWatchRUMAppLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatchrum.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatchrum.ListAppMonitorsInput{}

	for {
		output, err := svc.ListAppMonitors(params)
		if err != nil {
			return nil, err
		}

		for _, appEntry := range output.AppMonitorSummaries {
			resources = append(resources, &CloudWatchRumApp{
				svc:            svc,
				appMonitorName: appEntry.Name,
				id:             appEntry.Id,
				state:          appEntry.State,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CloudWatchRumApp struct {
	svc            *cloudwatchrum.CloudWatchRUM
	appMonitorName *string
	id             *string
	state          *string
}

func (f *CloudWatchRumApp) Remove(_ context.Context) error {
	_, err := f.svc.DeleteAppMonitor(&cloudwatchrum.DeleteAppMonitorInput{
		Name: f.appMonitorName,
	})

	return err
}

func (f *CloudWatchRumApp) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", *f.appMonitorName)
	properties.Set("ID", *f.id)
	properties.Set("State", *f.state)

	return properties
}

func (f *CloudWatchRumApp) String() string {
	return *f.appMonitorName
}
