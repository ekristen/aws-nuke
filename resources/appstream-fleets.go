package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appstream" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
)

const AppStreamFleetResource = "AppStreamFleet"

func init() {
	registry.Register(&registry.Registration{
		Name:     AppStreamFleetResource,
		Scope:    nuke.Account,
		Resource: &AppStreamFleet{},
		Lister:   &AppStreamFleetLister{},
	})
}

type AppStreamFleetLister struct{}

func (l *AppStreamFleetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := appstream.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &appstream.DescribeFleetsInput{}

	for {
		output, err := svc.DescribeFleets(params)
		if err != nil {
			return nil, err
		}

		for _, fleet := range output.Fleets {
			resources = append(resources, &AppStreamFleet{
				svc:  svc,
				name: fleet.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type AppStreamFleet struct {
	svc  *appstream.AppStream
	name *string
}

func (f *AppStreamFleet) Remove(_ context.Context) error {
	_, err := f.svc.StopFleet(&appstream.StopFleetInput{
		Name: f.name,
	})

	if err != nil {
		return err
	}

	_, err = f.svc.DeleteFleet(&appstream.DeleteFleetInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamFleet) String() string {
	return *f.name
}
