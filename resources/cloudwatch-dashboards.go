package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudWatchDashboardResource = "CloudWatchDashboard"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudWatchDashboardResource,
		Scope:  nuke.Account,
		Lister: &CloudWatchDashboardLister{},
	})
}

type CloudWatchDashboardLister struct{}

func (l *CloudWatchDashboardLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudwatch.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &cloudwatch.ListDashboardsInput{}

	for {
		output, err := svc.ListDashboards(params)
		if err != nil {
			return nil, err
		}

		for _, dashboardEntry := range output.DashboardEntries {
			resources = append(resources, &CloudWatchDashboard{
				svc:           svc,
				dashboardName: dashboardEntry.DashboardName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type CloudWatchDashboard struct {
	svc           *cloudwatch.CloudWatch
	dashboardName *string
}

func (f *CloudWatchDashboard) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDashboards(&cloudwatch.DeleteDashboardsInput{
		DashboardNames: []*string{f.dashboardName},
	})

	return err
}

func (f *CloudWatchDashboard) String() string {
	return *f.dashboardName
}
