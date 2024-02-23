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

const WAFRegionalRegexMatchTupleResource = "WAFRegionalRegexMatchTuple"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFRegionalRegexMatchTupleResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalRegexMatchTupleLister{},
	})
}

type WAFRegionalRegexMatchTupleLister struct{}

func (l *WAFRegionalRegexMatchTupleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			regexMatchSet, err := svc.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
				RegexMatchSetId: set.RegexMatchSetId,
			})
			if err != nil {
				return nil, err
			}

			for _, tuple := range regexMatchSet.RegexMatchSet.RegexMatchTuples {
				resources = append(resources, &WAFRegionalRegexMatchTuple{
					svc:        svc,
					matchSetID: set.RegexMatchSetId,
					tuple:      tuple,
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

type WAFRegionalRegexMatchTuple struct {
	svc        *wafregional.WAFRegional
	matchSetID *string
	tuple      *waf.RegexMatchTuple
}

func (r *WAFRegionalRegexMatchTuple) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.UpdateRegexMatchSet(&waf.UpdateRegexMatchSetInput{
		ChangeToken:     tokenOutput.ChangeToken,
		RegexMatchSetId: r.matchSetID,
		Updates: []*waf.RegexMatchSetUpdate{
			{
				Action:          aws.String("DELETE"),
				RegexMatchTuple: r.tuple,
			},
		},
	})

	return err
}

func (r *WAFRegionalRegexMatchTuple) Properties() types.Properties {
	return types.NewProperties().
		Set("RegexMatchSetID", r.matchSetID).
		Set("FieldToMatchType", r.tuple.FieldToMatch.Type).
		Set("FieldToMatchData", r.tuple.FieldToMatch.Data).
		Set("TextTransformation", r.tuple.TextTransformation)
}
