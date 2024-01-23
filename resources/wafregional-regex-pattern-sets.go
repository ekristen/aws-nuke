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

const WAFRegionalRegexPatternSetResource = "WAFRegionalRegexPatternSet"

func init() {
	resource.Register(&resource.Registration{
		Name:   WAFRegionalRegexPatternSetResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalRegexPatternSetLister{},
	})
}

type WAFRegionalRegexPatternSetLister struct{}

func (l *WAFRegionalRegexPatternSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafregional.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &waf.ListRegexPatternSetsInput{
		Limit: aws.Int64(50),
	}

	for {
		resp, err := svc.ListRegexPatternSets(params)
		if err != nil {
			return nil, err
		}

		for _, set := range resp.RegexPatternSets {
			resources = append(resources, &WAFRegionalRegexPatternSet{
				svc:  svc,
				id:   set.RegexPatternSetId,
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

type WAFRegionalRegexPatternSet struct {
	svc  *wafregional.WAFRegional
	id   *string
	name *string
}

func (r *WAFRegionalRegexPatternSet) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.DeleteRegexPatternSet(&waf.DeleteRegexPatternSetInput{
		RegexPatternSetId: r.id,
		ChangeToken:       tokenOutput.ChangeToken,
	})

	return err
}

func (r *WAFRegionalRegexPatternSet) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", r.id).
		Set("Name", r.name)
}

func (r *WAFRegionalRegexPatternSet) String() string {
	return *r.id
}
