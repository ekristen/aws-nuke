package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftFleetResource = "GameLiftFleet"

func init() {
	registry.Register(&registry.Registration{
		Name:   GameLiftFleetResource,
		Scope:  nuke.Account,
		Lister: &GameLiftFleetLister{},
	})
}

type GameLiftFleetLister struct{}

func (l *GameLiftFleetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := gamelift.New(opts.Session)

	params := &gamelift.ListFleetsInput{}

	for {
		resp, err := svc.ListFleets(params)
		if err != nil {
			return nil, err
		}

		for _, fleetID := range resp.FleetIds {
			fleet := &GameLiftFleet{
				svc:     svc,
				FleetID: fleetID,
			}
			resources = append(resources, fleet)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type GameLiftFleet struct {
	svc     *gamelift.GameLift
	FleetID *string
}

func (r *GameLiftFleet) Remove(_ context.Context) error {
	params := &gamelift.DeleteFleetInput{
		FleetId: r.FleetID,
	}

	_, err := r.svc.DeleteFleet(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLiftFleet) String() string {
	return *r.FleetID
}

func (r *GameLiftFleet) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
