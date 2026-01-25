package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/waf"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/wafregional" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRegionalRateBasedRulePredicateResource = "WAFRegionalRateBasedRulePredicate"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRegionalRateBasedRulePredicateResource,
		Scope:    nuke.Account,
		Resource: &WAFRegionalRateBasedRulePredicate{},
		Lister:   &WAFRegionalRateBasedRulePredicateLister{},
	})
}

type WAFRegionalRateBasedRulePredicateLister struct{}

func (l *WAFRegionalRateBasedRulePredicateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListRateBasedRulesInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListRateBasedRules(params)
		if err != nil {
			return nil, err
		}

		for _, rule := range resp.Rules {
			details, err := svc.GetRateBasedRule(&waf.GetRateBasedRuleInput{
				RuleId: rule.RuleId,
			})
			if err != nil {
				return nil, err
			}

			for _, predicate := range details.Rule.MatchPredicates {
				resources = append(resources, &WAFRegionalRateBasedRulePredicate{
					svc:       svc,
					ruleID:    rule.RuleId,
					rateLimit: details.Rule.RateLimit,
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

type WAFRegionalRateBasedRulePredicate struct {
	svc       *wafregional.WAFRegional
	ruleID    *string
	predicate *waf.Predicate
	rateLimit *int64
}

func (r *WAFRegionalRateBasedRulePredicate) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.UpdateRateBasedRule(&waf.UpdateRateBasedRuleInput{
		ChangeToken: tokenOutput.ChangeToken,
		RuleId:      r.ruleID,
		RateLimit:   r.rateLimit,
		Updates: []*waf.RuleUpdate{
			{
				Action:    aws.String("DELETE"),
				Predicate: r.predicate,
			},
		},
	})

	return err
}

func (r *WAFRegionalRateBasedRulePredicate) Properties() types.Properties {
	return types.NewProperties().
		Set("RuleID", r.ruleID).
		Set("Type", r.predicate.Type).
		Set("Negated", r.predicate.Negated).
		Set("DataID", r.predicate.DataId)
}

func (r *WAFRegionalRateBasedRulePredicate) String() string {
	return *r.ruleID
}
