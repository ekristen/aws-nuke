package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/gamelift"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

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

	svc := gamelift.New(opts.Session)

	resp, err := svc.DescribeMatchmakingRuleSets(&gamelift.DescribeMatchmakingRuleSetsInput{})
	if err != nil {
		return nil, err
	}

	rules := make([]resource.Resource, 0)
	for _, ruleSet := range resp.RuleSets {
		q := &GameLiftMatchmakingRuleSet{
			svc:  svc,
			Name: ruleSet.RuleSetName,
		}
		rules = append(rules, q)
	}

	return rules, nil
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

func (r *GameLiftMatchmakingRuleSet) String() string {
	return *r.Name
}
