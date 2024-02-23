package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/appstream"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type AppStreamFleetState struct {
	svc   *appstream.AppStream
	name  *string
	state *string
}

const AppStreamFleetStateResource = "AppStreamFleetState"

func init() {
	registry.Register(&registry.Registration{
		Name:   AppStreamFleetStateResource,
		Scope:  nuke.Account,
		Lister: &AppStreamFleetStateLister{},
	})
}

type AppStreamFleetStateLister struct{}

func (l *AppStreamFleetStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &AppStreamFleetState{
				svc:   svc,
				name:  fleet.Name,
				state: fleet.State,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

func (f *AppStreamFleetState) Remove(_ context.Context) error {
	_, err := f.svc.StopFleet(&appstream.StopFleetInput{
		Name: f.name,
	})

	return err
}

func (f *AppStreamFleetState) String() string {
	return *f.name
}

func (f *AppStreamFleetState) Filter() error {
	if ptr.ToString(f.state) == appstream.FleetStateStopped {
		return fmt.Errorf("already stopped")
	} else if ptr.ToString(f.state) == "DELETING" {
		return fmt.Errorf("already being deleted")
	}

	return nil
}
