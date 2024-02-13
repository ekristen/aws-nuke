package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const WAFRegionalRuleGroupResource = "WAFRegionalRuleGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFRegionalRuleGroupResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalRuleGroupLister{},
	})
}

type WAFRegionalRuleGroupLister struct{}

func (l *WAFRegionalRuleGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListRuleGroupsInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListRuleGroups(params)
		if err != nil {
			return nil, err
		}

		for _, rule := range resp.RuleGroups {
			resources = append(resources, &WAFRegionalRuleGroup{
				svc:  svc,
				ID:   rule.RuleGroupId,
				name: rule.Name,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalRuleGroup struct {
	svc  *wafregional.WAFRegional
	ID   *string
	name *string
}

func (f *WAFRegionalRuleGroup) Remove(_ context.Context) error {
	tokenOutput, err := f.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteRuleGroup(&waf.DeleteRuleGroupInput{
		RuleGroupId: f.ID,
		ChangeToken: tokenOutput.ChangeToken,
	})

	return err
}

func (f *WAFRegionalRuleGroup) String() string {
	return *f.ID
}

func (f *WAFRegionalRuleGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("ID", f.ID).
		Set("Name", f.name)
	return properties
}
