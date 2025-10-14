package resources

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/service/gamelift" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftMatchmakingConfigurationResource = "GameLiftMatchmakingConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:     GameLiftMatchmakingConfigurationResource,
		Scope:    nuke.Account,
		Resource: &GameLiftMatchmakingConfiguration{},
		Lister:   &GameLiftMatchmakingConfigurationLister{},
	})
}

type GameLiftMatchmakingConfigurationLister struct {
	GameLift
}

func (l *GameLiftMatchmakingConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		opts.Logger.
			WithField("resource", GameLiftMatchmakingConfigurationResource).
			WithField("region", opts.Region.Name).
			Debug("region not supported")
		return resources, nil
	}

	svc := gamelift.New(opts.Session)

	params := &gamelift.DescribeMatchmakingConfigurationsInput{}

	for {
		resp, err := svc.DescribeMatchmakingConfigurations(params)
		if err != nil {
			var unsupportedRegionException *gamelift.UnsupportedRegionException
			if errors.As(err, &unsupportedRegionException) {
				return resources, nil
			}
			return resources, err
		}

		for _, config := range resp.Configurations {
			q := &GameLiftMatchmakingConfiguration{
				svc:          svc,
				Name:         config.Name,
				CreationTime: config.CreationTime,
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

type GameLiftMatchmakingConfiguration struct {
	svc          *gamelift.GameLift
	Name         *string
	CreationTime *time.Time
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

func (r *GameLiftMatchmakingConfiguration) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *GameLiftMatchmakingConfiguration) String() string {
	return *r.Name
}
