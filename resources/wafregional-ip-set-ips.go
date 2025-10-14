package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/waf"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/wafregional" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFRegionalIPSetIPResource = "WAFRegionalIPSetIP"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRegionalIPSetIPResource,
		Scope:    nuke.Account,
		Resource: &WAFRegionalIPSetIP{},
		Lister:   &WAFRegionalIPSetIPLister{},
	})
}

type WAFRegionalIPSetIPLister struct{}

func (l *WAFRegionalIPSetIPLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			details, err := svc.GetIPSet(&waf.GetIPSetInput{
				IPSetId: set.IPSetId,
			})
			if err != nil {
				return nil, err
			}

			for _, descriptor := range details.IPSet.IPSetDescriptors {
				resources = append(resources, &WAFRegionalIPSetIP{
					svc:        svc,
					ipSetID:    set.IPSetId,
					descriptor: descriptor,
				})
			}
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFRegionalIPSetIP struct {
	svc        *wafregional.WAFRegional
	ipSetID    *string
	descriptor *waf.IPSetDescriptor
}

func (r *WAFRegionalIPSetIP) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.UpdateIPSet(&waf.UpdateIPSetInput{
		ChangeToken: tokenOutput.ChangeToken,
		IPSetId:     r.ipSetID,
		Updates: []*waf.IPSetUpdate{
			{
				Action:          aws.String("DELETE"),
				IPSetDescriptor: r.descriptor,
			},
		},
	})

	return err
}

func (r *WAFRegionalIPSetIP) Properties() types.Properties {
	return types.NewProperties().
		Set("IPSetID", r.ipSetID).
		Set("Type", r.descriptor.Type).
		Set("Value", r.descriptor.Value)
}
