package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const WAFRegionalByteMatchSetResource = "WAFRegionalByteMatchSet"

func init() {
	resource.Register(&resource.Registration{
		Name:   WAFRegionalByteMatchSetResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalByteMatchSetLister{},
	})
}

type WAFRegionalByteMatchSetLister struct{}

func (l *WAFRegionalByteMatchSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListByteMatchSetsInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListByteMatchSets(params)
		if err != nil {
			return nil, err
		}

		for _, set := range resp.ByteMatchSets {
			resources = append(resources, &WAFRegionalByteMatchSet{
				svc:  svc,
				id:   set.ByteMatchSetId,
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

type WAFRegionalByteMatchSet struct {
	svc  *wafregional.WAFRegional
	id   *string
	name *string
}

func (r *WAFRegionalByteMatchSet) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.DeleteByteMatchSet(&waf.DeleteByteMatchSetInput{
		ByteMatchSetId: r.id,
		ChangeToken:    tokenOutput.ChangeToken,
	})

	return err
}

func (r *WAFRegionalByteMatchSet) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", r.id).
		Set("Name", r.name)
}

func (r *WAFRegionalByteMatchSet) String() string {
	return *r.id
}
