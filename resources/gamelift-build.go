package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/gamelift" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftBuildResource = "GameLiftBuild"

func init() {
	registry.Register(&registry.Registration{
		Name:     GameLiftBuildResource,
		Scope:    nuke.Account,
		Resource: &GameLiftBuild{},
		Lister:   &GameLiftBuildLister{},
	})
}

type GameLiftBuildLister struct {
	GameLift
}

func (l *GameLiftBuildLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		opts.Logger.
			WithField("resource", GameLiftBuildResource).
			WithField("region", opts.Region.Name).
			Debug("region not supported")
		return resources, nil
	}

	svc := gamelift.New(opts.Session)

	params := &gamelift.ListBuildsInput{}

	for {
		resp, err := svc.ListBuilds(params)
		if err != nil {
			return nil, err
		}

		for _, build := range resp.Builds {
			resources = append(resources, &GameLiftBuild{
				svc:          svc,
				BuildID:      build.BuildId,
				Name:         build.Name,
				Status:       build.Status,
				Version:      build.Version,
				CreationDate: build.CreationTime,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type GameLiftBuild struct {
	svc          *gamelift.GameLift
	BuildID      *string
	Name         *string
	Status       *string
	Version      *string
	CreationDate *time.Time
}

func (r *GameLiftBuild) Remove(_ context.Context) error {
	params := &gamelift.DeleteBuildInput{
		BuildId: r.BuildID,
	}

	_, err := r.svc.DeleteBuild(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLiftBuild) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *GameLiftBuild) String() string {
	return *r.BuildID
}
