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

const WAFRegionalRegexMatchSetResource = "WAFRegionalRegexMatchSet" //nolint:gosec,nolintlint

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFRegionalRegexMatchSetResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalRegexMatchSetLister{},
	})
}

type WAFRegionalRegexMatchSetLister struct{}

func (l *WAFRegionalRegexMatchSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListRegexMatchSetsInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListRegexMatchSets(params)
		if err != nil {
			return nil, err
		}

		for _, set := range resp.RegexMatchSets {
			resources = append(resources, &WAFRegionalRegexMatchSet{
				svc:  svc,
				id:   set.RegexMatchSetId,
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

type WAFRegionalRegexMatchSet struct {
	svc  *wafregional.WAFRegional
	id   *string
	name *string
}

func (r *WAFRegionalRegexMatchSet) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.DeleteRegexMatchSet(&waf.DeleteRegexMatchSetInput{
		RegexMatchSetId: r.id,
		ChangeToken:     tokenOutput.ChangeToken,
	})

	return err
}

func (r *WAFRegionalRegexMatchSet) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", r.id).
		Set("Name", r.name)
}

func (r *WAFRegionalRegexMatchSet) String() string {
	return *r.id
}
