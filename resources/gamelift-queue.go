package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftQueueResource = "GameLiftQueue"

func init() {
	registry.Register(&registry.Registration{
		Name:   GameLiftQueueResource,
		Scope:  nuke.Account,
		Lister: &GameLiftQueueLister{},
	})
}

type GameLiftQueueLister struct{}

func (l *GameLiftQueueLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gamelift.New(opts.Session)

	resp, err := svc.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{})
	if err != nil {
		return nil, err
	}

	queues := make([]resource.Resource, 0)
	for _, queue := range resp.GameSessionQueues {
		q := &GameLiftQueue{
			svc:  svc,
			Name: queue.Name,
		}
		queues = append(queues, q)
	}

	return queues, nil
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

func (r *GameLiftQueue) String() string {
	return *r.Name
}
