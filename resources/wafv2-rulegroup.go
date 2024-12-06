package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/wafv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFv2RuleGroupResource = "WAFv2RuleGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFv2RuleGroupResource,
		Scope:    nuke.Account,
		Resource: &WAFv2RuleGroup{},
		Lister:   &WAFv2RuleGroupLister{},
	})
}

type WAFv2RuleGroupLister struct{}

func (l *WAFv2RuleGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &wafv2.ListRuleGroupsInput{
		Limit: aws.Int64(50),
		Scope: aws.String("REGIONAL"),
	}

	output, err := getRuleGroups(svc, params)
	if err != nil {
		return []resource.Resource{}, err
	}

	resources = append(resources, output...)

	if *opts.Session.Config.Region == endpoints.UsEast1RegionID {
		params.Scope = aws.String("CLOUDFRONT")

		output, err := getRuleGroups(svc, params)
		if err != nil {
			return []resource.Resource{}, err
		}

		resources = append(resources, output...)
	}

	return resources, nil
}

func getRuleGroups(svc *wafv2.WAFV2, params *wafv2.ListRuleGroupsInput) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.ListRuleGroups(params)
		if err != nil {
			return nil, err
		}

		for _, webACL := range resp.RuleGroups {
			resources = append(resources, &WAFv2RuleGroup{
				svc:       svc,
				ID:        webACL.Id,
				name:      webACL.Name,
				lockToken: webACL.LockToken,
				scope:     params.Scope,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}
	return resources, nil
}

type WAFv2RuleGroup struct {
	svc       *wafv2.WAFV2
	ID        *string
	name      *string
	lockToken *string
	scope     *string
}

func (f *WAFv2RuleGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteRuleGroup(&wafv2.DeleteRuleGroupInput{
		Id:        f.ID,
		Name:      f.name,
		Scope:     f.scope,
		LockToken: f.lockToken,
	})

	return err
}

func (f *WAFv2RuleGroup) String() string {
	return *f.ID
}

func (f *WAFv2RuleGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("ID", f.ID).
		Set("Name", f.name).
		Set("Scope", f.scope)
	return properties
}
