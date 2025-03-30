package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"go.uber.org/ratelimit"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRegionalRuleResource = "WAFRegionalRule"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRegionalRuleResource,
		Scope:    nuke.Account,
		Resource: &WAFRegionalRule{},
		Lister:   &WAFRegionalRuleLister{},
	})
}

type WAFRegionalRuleLister struct{}

func (l *WAFRegionalRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	listRl := ratelimit.New(15)
	getRl := ratelimit.New(10)

	params := &waf.ListRulesInput{
		Limit: aws.Int64(50),
	}

	for {
		listRl.Take() // Wait for ListRules rate limiter

		resp, err := svc.ListRules(params)
		if err != nil {
			return nil, err
		}

		for _, rule := range resp.Rules {
			getRl.Take() // Wait for GetRule rate limiter

			ruleResp, err := svc.GetRule(&waf.GetRuleInput{
				RuleId: rule.RuleId,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &WAFRegionalRule{
				svc:  svc,
				ID:   rule.RuleId,
				name: rule.Name,
				rule: ruleResp.Rule,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalRule struct {
	svc  *wafregional.WAFRegional
	ID   *string
	name *string
	rule *waf.Rule
}

func (f *WAFRegionalRule) Remove(_ context.Context) error {
	tokenOutput, err := f.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	var ruleUpdates []*waf.RuleUpdate
	for _, predicate := range f.rule.Predicates {
		ruleUpdates = append(ruleUpdates, &waf.RuleUpdate{
			Action:    aws.String(waf.ChangeActionDelete),
			Predicate: predicate,
		})
	}

	_, err = f.svc.UpdateRule(&waf.UpdateRuleInput{
		ChangeToken: tokenOutput.ChangeToken,
		RuleId:      f.ID,
		Updates:     ruleUpdates,
	})

	if err != nil {
		return err
	}

	tokenOutput, err = f.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteRule(&waf.DeleteRuleInput{
		RuleId:      f.ID,
		ChangeToken: tokenOutput.ChangeToken,
	})

	return err
}

func (f *WAFRegionalRule) String() string {
	return *f.ID
}

func (f *WAFRegionalRule) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("ID", f.ID).
		Set("Name", f.name)
	return properties
}
