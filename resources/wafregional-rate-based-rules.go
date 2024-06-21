package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRegionalRateBasedRuleResource = "WAFRegionalRateBasedRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFRegionalRateBasedRuleResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalRateBasedRuleLister{},
	})
}

type WAFRegionalRateBasedRuleLister struct{}

func (l *WAFRegionalRateBasedRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &WAFRegionalRateBasedRule{
				svc: svc,
				ID:  rule.RuleId,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalRateBasedRule struct {
	svc *wafregional.WAFRegional
	ID  *string
}

func (f *WAFRegionalRateBasedRule) Remove(_ context.Context) error {
	tokenOutput, err := f.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteRateBasedRule(&waf.DeleteRateBasedRuleInput{
		RuleId:      f.ID,
		ChangeToken: tokenOutput.ChangeToken,
	})

	return err
}

func (f *WAFRegionalRateBasedRule) String() string {
	return *f.ID
}
