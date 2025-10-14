package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/opsworks" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const OpsWorksAppResource = "OpsWorksApp"

func init() {
	registry.Register(&registry.Registration{
		Name:     OpsWorksAppResource,
		Scope:    nuke.Account,
		Resource: &OpsWorksApp{},
		Lister:   &OpsWorksAppLister{},
	})
}

type OpsWorksAppLister struct{}

func (l *OpsWorksAppLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := opsworks.New(opts.Session)
	resources := make([]resource.Resource, 0)

	stackParams := &opsworks.DescribeStacksInput{}

	resp, err := svc.DescribeStacks(stackParams)
	if err != nil {
		return nil, err
	}

	appsParams := &opsworks.DescribeAppsInput{}

	for _, stack := range resp.Stacks {
		appsParams.StackId = stack.StackId
		output, err := svc.DescribeApps(appsParams)
		if err != nil {
			return nil, err
		}

		for _, app := range output.Apps {
			resources = append(resources, &OpsWorksApp{
				svc: svc,
				ID:  app.AppId,
			})
		}
	}

	return resources, nil
}

type OpsWorksApp struct {
	svc *opsworks.OpsWorks
	ID  *string
}

func (f *OpsWorksApp) Remove(_ context.Context) error {
	_, err := f.svc.DeleteApp(&opsworks.DeleteAppInput{
		AppId: f.ID,
	})

	return err
}

func (f *OpsWorksApp) String() string {
	return *f.ID
}
