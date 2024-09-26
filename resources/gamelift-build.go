package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftBuildResource = "GameLiftBuild"

func init() {
	registry.Register(&registry.Registration{
		Name:   GameLiftBuildResource,
		Scope:  nuke.Account,
		Lister: &GameLiftBuildLister{},
	})
}

type GameLiftBuildLister struct{}

func (l *GameLiftBuildLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := gamelift.New(opts.Session)

	resp, err := svc.ListBuilds(&gamelift.ListBuildsInput{})
	if err != nil {
		return nil, err
	}

	builds := make([]resource.Resource, 0)
	for _, build := range resp.Builds {
		builds = append(builds, &GameLiftBuild{
			svc:     svc,
			BuildID: build.BuildId,
		})
	}

	return builds, nil
}

type GameLiftBuild struct {
	svc     *gamelift.GameLift
	BuildID *string
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

func (r *GameLiftBuild) String() string {
	return *r.BuildID
}
