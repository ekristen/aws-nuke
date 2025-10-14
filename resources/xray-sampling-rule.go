package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/xray" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const XRaySamplingRuleResource = "XRaySamplingRule"

func init() {
	registry.Register(&registry.Registration{
		Name:     XRaySamplingRuleResource,
		Scope:    nuke.Account,
		Resource: &XRaySamplingRule{},
		Lister:   &XRaySamplingRuleLister{},
	})
}

type XRaySamplingRuleLister struct{}

func (l *XRaySamplingRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := xray.New(opts.Session)
	resources := make([]resource.Resource, 0)

	var xraySamplingRules []*xray.SamplingRule
	err := svc.GetSamplingRulesPages(
		&xray.GetSamplingRulesInput{},
		func(page *xray.GetSamplingRulesOutput, lastPage bool) bool {
			for _, rule := range page.SamplingRuleRecords {
				if *rule.SamplingRule.RuleName != "Default" {
					xraySamplingRules = append(xraySamplingRules, rule.SamplingRule)
				}
			}
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	for _, rule := range xraySamplingRules {
		resources = append(resources, &XRaySamplingRule{
			svc:      svc,
			ruleName: rule.RuleName,
			ruleARN:  rule.RuleARN,
		})
	}

	return resources, nil
}

type XRaySamplingRule struct {
	svc      *xray.XRay
	ruleName *string
	ruleARN  *string
}

func (f *XRaySamplingRule) Remove(_ context.Context) error {
	_, err := f.svc.DeleteSamplingRule(&xray.DeleteSamplingRuleInput{
		RuleARN: f.ruleARN, // Specify ruleARN or ruleName, not both
	})

	return err
}

func (f *XRaySamplingRule) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("RuleName", f.ruleName).
		Set("RuleARN", f.ruleARN)

	return properties
}
