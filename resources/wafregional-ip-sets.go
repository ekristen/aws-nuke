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

const WAFRegionalIPSetResource = "WAFRegionalIPSet"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRegionalIPSetResource,
		Scope:    nuke.Account,
		Resource: &WAFRegionalIPSet{},
		Lister:   &WAFRegionalIPSetLister{},
	})
}

type WAFRegionalIPSetLister struct{}

func (l *WAFRegionalIPSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListIPSetsInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListIPSets(params)
		if err != nil {
			return nil, err
		}

		for _, set := range resp.IPSets {
			resources = append(resources, &WAFRegionalIPSet{
				svc:  svc,
				id:   set.IPSetId,
				name: set.Name,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalIPSet struct {
	svc  *wafregional.WAFRegional
	id   *string
	name *string
}

func (r *WAFRegionalIPSet) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.DeleteIPSet(&waf.DeleteIPSetInput{
		IPSetId:     r.id,
		ChangeToken: tokenOutput.ChangeToken,
	})

	return err
}

func (r *WAFRegionalIPSet) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", r.id).
		Set("Name", r.name)
}

func (r *WAFRegionalIPSet) String() string {
	return *r.id
}
