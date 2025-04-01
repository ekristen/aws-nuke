package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"go.uber.org/ratelimit"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRuleResource = "WAFRule"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRuleResource,
		Scope:    nuke.Account,
		Resource: &WAFRule{},
		Lister:   &WAFRuleLister{},
	})
}

type WAFRuleLister struct{}

func (l *WAFRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := waf.New(opts.Session)
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

			resources = append(resources, &WAFRule{
				svc:  svc,
				ID:   rule.RuleId,
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

type WAFRule struct {
	svc  *waf.WAF
	ID   *string
	rule *waf.Rule
}

func (f *WAFRule) Remove(_ context.Context) error {
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

func (f *WAFRule) String() string {
	return *f.ID
}

func (f *WAFRule) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("ID", f.ID).
		Set("Name", f.rule.Name)

	return properties
}
