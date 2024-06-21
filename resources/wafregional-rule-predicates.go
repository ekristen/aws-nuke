package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRegionalRulePredicateResource = "WAFRegionalRulePredicate"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFRegionalRulePredicateResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalRulePredicateLister{},
	})
}

type WAFRegionalRulePredicateLister struct{}

func (l *WAFRegionalRulePredicateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListRulesInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListRules(params)
		if err != nil {
			return nil, err
		}

		for _, rule := range resp.Rules {
			details, err := svc.GetRule(&waf.GetRuleInput{
				RuleId: rule.RuleId,
			})
			if err != nil {
				return nil, err
			}

			for _, predicate := range details.Rule.Predicates {
				resources = append(resources, &WAFRegionalRulePredicate{
					svc:       svc,
					ruleID:    rule.RuleId,
					predicate: predicate,
				})
			}
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalRulePredicate struct {
	svc       *wafregional.WAFRegional
	ruleID    *string
	predicate *waf.Predicate
}

func (r *WAFRegionalRulePredicate) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.UpdateRule(&waf.UpdateRuleInput{
		ChangeToken: tokenOutput.ChangeToken,
		RuleId:      r.ruleID,
		Updates: []*waf.RuleUpdate{
			{
				Action:    aws.String("DELETE"),
				Predicate: r.predicate,
			},
		},
	})

	return err
}

func (r *WAFRegionalRulePredicate) Properties() types.Properties {
	return types.NewProperties().
		Set("RuleID", r.ruleID).
		Set("Type", r.predicate.Type).
		Set("Negated", r.predicate.Negated).
		Set("DataID", r.predicate.DataId)
}

func (r *WAFRegionalRulePredicate) String() string {
	return *r.predicate.DataId
}
