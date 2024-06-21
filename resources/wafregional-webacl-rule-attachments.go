package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRegionalWebACLRuleAttachmentResource = "WAFRegionalWebACLRuleAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFRegionalWebACLRuleAttachmentResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalWebACLRuleAttachmentLister{},
	})
}

type WAFRegionalWebACLRuleAttachmentLister struct{}

func (l *WAFRegionalWebACLRuleAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var webACLs []*waf.WebACLSummary

	params := &waf.ListWebACLsInput{
		Limit: aws.Int64(50),
	}

	// List All Web ACL's
	for {
		resp, err := svc.ListWebACLs(params)
		if err != nil {
			return nil, err
		}

		webACLs = append(webACLs, resp.WebACLs...)

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	webACLParams := &waf.GetWebACLInput{}

	for _, webACL := range webACLs {
		webACLParams.WebACLId = webACL.WebACLId

		resp, err := svc.GetWebACL(webACLParams)
		if err != nil {
			return nil, err
		}

		for _, webACLRule := range resp.WebACL.Rules {
			resources = append(resources, &WAFRegionalWebACLRuleAttachment{
				svc:           svc,
				webACLID:      webACL.WebACLId,
				activatedRule: webACLRule,
			})
		}
	}

	return resources, nil
}

type WAFRegionalWebACLRuleAttachment struct {
	svc           *wafregional.WAFRegional
	webACLID      *string
	activatedRule *waf.ActivatedRule
}

func (f *WAFRegionalWebACLRuleAttachment) Remove(_ context.Context) error {
	tokenOutput, err := f.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	webACLUpdate := &waf.WebACLUpdate{
		Action:        aws.String("DELETE"),
		ActivatedRule: f.activatedRule,
	}

	_, err = f.svc.UpdateWebACL(&waf.UpdateWebACLInput{
		WebACLId:    f.webACLID,
		ChangeToken: tokenOutput.ChangeToken,
		Updates:     []*waf.WebACLUpdate{webACLUpdate},
	})

	return err
}

func (f *WAFRegionalWebACLRuleAttachment) String() string {
	return fmt.Sprintf("%s -> %s", *f.webACLID, *f.activatedRule.RuleId)
}
