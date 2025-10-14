package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftQueueResource = "GameLiftQueue"

func init() {
	registry.Register(&registry.Registration{
		Name:     GameLiftQueueResource,
		Scope:    nuke.Account,
		Resource: &GameLiftQueue{},
		Lister:   &GameLiftQueueLister{},
	})
}

type GameLiftQueueLister struct {
	GameLift
}

func (l *GameLiftQueueLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		opts.Logger.
			WithField("resource", GameLiftQueueResource).
			WithField("region", opts.Region.Name).
			Debug("region not supported")
		return resources, nil
	}

	svc := gamelift.New(opts.Session)

	params := &gamelift.DescribeGameSessionQueuesInput{}

	for {
		resp, err := svc.DescribeGameSessionQueues(params)
		if err != nil {
			return nil, err
		}

		for _, queue := range resp.GameSessionQueues {
			q := &GameLiftQueue{
				svc:  svc,
				Name: queue.Name,
			}
			resources = append(resources, q)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type GameLiftQueue struct {
	svc  *gamelift.GameLift
	Name *string
}

func (r *GameLiftQueue) Remove(_ context.Context) error {
	params := &gamelift.DeleteGameSessionQueueInput{
		Name: r.Name,
	}

	_, err := r.svc.DeleteGameSessionQueue(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLiftQueue) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *GameLiftQueue) String() string {
	return *r.Name
}
