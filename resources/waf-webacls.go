package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const WAFWebACLResource = "WAFWebACL"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFWebACLResource,
		Scope:  nuke.Account,
		Lister: &WAFWebACLLister{},
	})
}

type WAFWebACLLister struct{}

func (l *WAFWebACLLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := waf.New(opts.Session)
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
			resources = append(resources, &WAFWebACL{
				svc: svc,
				ID:  webACL.WebACLId,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFWebACL struct {
	svc *waf.WAF
	ID  *string
}

func (f *WAFWebACL) Remove(_ context.Context) error {
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

func (f *WAFWebACL) String() string {
	return *f.ID
}
