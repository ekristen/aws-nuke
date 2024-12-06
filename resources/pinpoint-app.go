package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/pinpoint"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const PinpointAppResource = "PinpointApp"

func init() {
	registry.Register(&registry.Registration{
		Name:     PinpointAppResource,
		Scope:    nuke.Account,
		Resource: &PinpointApp{},
		Lister:   &PinpointAppLister{},
	})
}

type PinpointAppLister struct{}

func (l *PinpointAppLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := pinpoint.New(opts.Session)

	params := &pinpoint.GetAppsInput{}

	for {
		resp, err := svc.GetApps(params)
		if err != nil {
			return nil, err
		}

		for _, appResponse := range resp.ApplicationsResponse.Item {
			resources = append(resources, &PinpointApp{
				svc:  svc,
				ID:   appResponse.Id,
				Name: appResponse.Name,
				Tags: appResponse.Tags,
			})
		}

		if resp.ApplicationsResponse.NextToken == nil {
			break
		}

		params.Token = resp.ApplicationsResponse.NextToken
	}

	return resources, nil
}

type PinpointApp struct {
	svc          *pinpoint.Pinpoint
	ID           *string
	Name         *string
	CreationDate *time.Time
	Tags         map[string]*string
}

func (r *PinpointApp) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *PinpointApp) Remove(_ context.Context) error {
	params := &pinpoint.DeleteAppInput{
		ApplicationId: r.ID,
	}

	_, err := r.svc.DeleteApp(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *PinpointApp) String() string {
	return *r.Name
}
