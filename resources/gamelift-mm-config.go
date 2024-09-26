package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftMatchmakingConfigurationResource = "GameLiftMatchmakingConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   GameLiftMatchmakingConfigurationResource,
		Scope:  nuke.Account,
		Lister: &GameLiftMatchmakingConfigurationLister{},
	})
}

type GameLiftMatchmakingConfigurationLister struct{}

func (l *GameLiftMatchmakingConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gamelift.New(opts.Session)

	resp, err := svc.DescribeMatchmakingConfigurations(&gamelift.DescribeMatchmakingConfigurationsInput{})
	if err != nil {
		return nil, err
	}

	configs := make([]resource.Resource, 0)
	for _, config := range resp.Configurations {
		q := &GameLiftMatchmakingConfiguration{
			svc:  svc,
			Name: config.Name,
		}
		configs = append(configs, q)
	}

	return configs, nil
}

type GameLiftMatchmakingConfiguration struct {
	svc  *gamelift.GameLift
	Name *string
}

func (r *GameLiftMatchmakingConfiguration) Remove(_ context.Context) error {
	params := &gamelift.DeleteMatchmakingConfigurationInput{
		Name: r.Name,
	}

	_, err := r.svc.DeleteMatchmakingConfiguration(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLiftMatchmakingConfiguration) String() string {
	return *r.Name
}
