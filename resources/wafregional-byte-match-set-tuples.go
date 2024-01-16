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

const WAFRegionalByteMatchSetIPResource = "WAFRegionalByteMatchSetIP"

func init() {
	resource.Register(resource.Registration{
		Name:   WAFRegionalByteMatchSetIPResource,
		Scope:  nuke.Account,
		Lister: &WAFRegionalByteMatchSetIPLister{},
	})
}

type WAFRegionalByteMatchSetIPLister struct{}

func (l *WAFRegionalByteMatchSetIPLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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

			details, err := svc.GetByteMatchSet(&waf.GetByteMatchSetInput{
				ByteMatchSetId: set.ByteMatchSetId,
			})
			if err != nil {
				return nil, err
			}

			for _, tuple := range details.ByteMatchSet.ByteMatchTuples {
				resources = append(resources, &WAFRegionalByteMatchSetIP{
					svc:        svc,
					matchSetID: set.ByteMatchSetId,
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

type WAFRegionalByteMatchSetIP struct {
	svc        *wafregional.WAFRegional
	matchSetID *string
	tuple      *waf.ByteMatchTuple
}

func (r *WAFRegionalByteMatchSetIP) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.UpdateByteMatchSet(&waf.UpdateByteMatchSetInput{
		ChangeToken:    tokenOutput.ChangeToken,
		ByteMatchSetId: r.matchSetID,
		Updates: []*waf.ByteMatchSetUpdate{
			&waf.ByteMatchSetUpdate{
				Action:         aws.String("DELETE"),
				ByteMatchTuple: r.tuple,
			},
		},
	})

	return err
}

func (r *WAFRegionalByteMatchSetIP) Properties() types.Properties {
	return types.NewProperties().
		Set("ByteMatchSetID", r.matchSetID).
		Set("FieldToMatchType", r.tuple.FieldToMatch.Type).
		Set("FieldToMatchData", r.tuple.FieldToMatch.Data).
		Set("TargetString", r.tuple.TargetString)
}
