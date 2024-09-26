package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/pinpoint"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const PinpointAppResource = "PinpointApp"

func init() {
	registry.Register(&registry.Registration{
		Name:   PinpointAppResource,
		Scope:  nuke.Account,
		Lister: &PinpointAppLister{},
	})
}

type PinpointAppLister struct{}

func (l *PinpointAppLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := pinpoint.New(opts.Session)

	resp, err := svc.GetApps(&pinpoint.GetAppsInput{})
	if err != nil {
		return nil, err
	}

	apps := make([]resource.Resource, 0)
	for _, appResponse := range resp.ApplicationsResponse.Item {
		apps = append(apps, &PinpointApp{
			svc: svc,
			ID:  appResponse.Id,
		})
	}

	return apps, nil
}

type PinpointApp struct {
	svc *pinpoint.Pinpoint
	ID  *string
}

func (p *PinpointApp) Remove(_ context.Context) error {
	params := &pinpoint.DeleteAppInput{
		ApplicationId: p.ID,
	}

	_, err := p.svc.DeleteApp(params)
	if err != nil {
		return err
	}

	return nil
}

func (p *PinpointApp) String() string {
	return *p.ID
}
