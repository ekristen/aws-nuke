package resources

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/service/gamelift" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftMatchmakingRuleSetResource = "GameLiftMatchmakingRuleSet"

func init() {
	registry.Register(&registry.Registration{
		Name:     GameLiftMatchmakingRuleSetResource,
		Scope:    nuke.Account,
		Resource: &GameLiftMatchmakingRuleSet{},
		Lister:   &GameLiftMatchmakingRuleSetLister{},
	})
}

type GameLiftMatchmakingRuleSetLister struct {
	GameLift
}

func (l *GameLiftMatchmakingRuleSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		opts.Logger.
			WithField("resource", GameLiftMatchmakingRuleSetResource).
			WithField("region", opts.Region.Name).
			Debug("region not supported")
		return resources, nil
	}

	svc := gamelift.New(opts.Session)

	params := &gamelift.DescribeMatchmakingRuleSetsInput{}

	for {
		resp, err := svc.DescribeMatchmakingRuleSets(params)
		if err != nil {
			var unsupportedRegionException *gamelift.UnsupportedRegionException
			if errors.As(err, &unsupportedRegionException) {
				return resources, nil
			}
			return resources, err
		}

		for _, ruleSet := range resp.RuleSets {
			q := &GameLiftMatchmakingRuleSet{
				svc:  svc,
				Name: ruleSet.RuleSetName,
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

type GameLiftMatchmakingRuleSet struct {
	svc  *gamelift.GameLift
	Name *string
}

func (r *GameLiftMatchmakingRuleSet) Remove(_ context.Context) error {
	params := &gamelift.DeleteMatchmakingRuleSetInput{
		Name: r.Name,
	}

	_, err := r.svc.DeleteMatchmakingRuleSet(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *GameLiftMatchmakingRuleSet) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *GameLiftMatchmakingRuleSet) String() string {
	return *r.Name
}
