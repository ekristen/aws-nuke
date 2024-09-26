package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GameLiftMatchmakingRuleSetResource = "GameLiftMatchmakingRuleSet"

func init() {
	registry.Register(&registry.Registration{
		Name:   GameLiftMatchmakingRuleSetResource,
		Scope:  nuke.Account,
		Lister: &GameLiftMatchmakingRuleSetLister{},
	})
}

type GameLiftMatchmakingRuleSetLister struct{}

func (l *GameLiftMatchmakingRuleSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	var resources []resource.Resource

	svc := gamelift.New(opts.Session)

	params := &gamelift.DescribeMatchmakingRuleSetsInput{}

	for {
		resp, err := svc.DescribeMatchmakingRuleSets(params)
		if err != nil {
			return nil, err
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
