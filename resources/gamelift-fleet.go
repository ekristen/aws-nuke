package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

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

	svc := gamelift.New(opts.Session)

	resp, err := svc.ListFleets(&gamelift.ListFleetsInput{})
	if err != nil {
		return nil, err
	}

	fleets := make([]resource.Resource, 0)
	for _, fleetId := range resp.FleetIds {
		fleet := &GameLiftFleet{
			svc:     svc,
			FleetId: *fleetId, // Dereference the fleetId pointer
		}
		fleets = append(fleets, fleet)
	}

	return fleets, nil
}

type GameLiftFleet struct {
	svc     *gamelift.GameLift
	FleetId string
}

func (r *GameLiftFleet) Remove(_ context.Context) error {
	params := &gamelift.DeleteFleetInput{
		FleetId: aws.String(r.FleetId),
	}

	_, err := r.svc.DeleteFleet(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLiftFleet) String() string {
	return r.FleetId
}
