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

const WAFRegionalWebACLResource = "WAFRegionalWebACL"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRegionalWebACLResource,
		Scope:    nuke.Account,
		Resource: &WAFRegionalWebACL{},
		Lister:   &WAFRegionalWebACLLister{},
	})
}

type WAFRegionalWebACLLister struct{}

func (l *WAFRegionalWebACLLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListWebACLsInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListWebACLs(params)
		if err != nil {
			return nil, err
		}

		for _, webACL := range resp.WebACLs {
			resources = append(resources, &WAFRegionalWebACL{
				svc:  svc,
				ID:   webACL.WebACLId,
				name: webACL.Name,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalWebACL struct {
	svc  *wafregional.WAFRegional
	ID   *string
	name *string
}

func (f *WAFRegionalWebACL) Remove(_ context.Context) error {
	tokenOutput, err := f.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteWebACL(&waf.DeleteWebACLInput{
		WebACLId:    f.ID,
		ChangeToken: tokenOutput.ChangeToken,
	})

	return err
}

func (f *WAFRegionalWebACL) String() string {
	return *f.ID
}

func (f *WAFRegionalWebACL) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("ID", f.ID).
		Set("Name", f.name)
	return properties
}
