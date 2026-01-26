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

const WAFRegionalRegexPatternStringResource = "WAFRegionalRegexPatternString"

func init() {
	registry.Register(&registry.Registration{
		Name:     WAFRegionalRegexPatternStringResource,
		Scope:    nuke.Account,
		Resource: &WAFRegionalRegexPatternString{},
		Lister:   &WAFRegionalRegexPatternStringLister{},
	})
}

type WAFRegionalRegexPatternStringLister struct{}

func (l *WAFRegionalRegexPatternStringLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			regexPatternSet, err := svc.GetRegexPatternSet(&waf.GetRegexPatternSetInput{
				RegexPatternSetId: set.RegexPatternSetId,
			})
			if err != nil {
				return nil, err
			}

			for _, patternString := range regexPatternSet.RegexPatternSet.RegexPatternStrings {
				resources = append(resources, &WAFRegionalRegexPatternString{
					svc:           svc,
					patternSetID:  set.RegexPatternSetId,
					patternString: patternString,
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

type WAFRegionalRegexPatternString struct {
	svc           *wafregional.WAFRegional
	patternSetID  *string
	patternString *string
}

func (r *WAFRegionalRegexPatternString) Remove(_ context.Context) error {
	tokenOutput, err := r.svc.GetChangeToken(&waf.GetChangeTokenInput{})
	if err != nil {
		return err
	}

	_, err = r.svc.UpdateRegexPatternSet(&waf.UpdateRegexPatternSetInput{
		ChangeToken:       tokenOutput.ChangeToken,
		RegexPatternSetId: r.patternSetID,
		Updates: []*waf.RegexPatternSetUpdate{
			{
				Action:             aws.String("DELETE"),
				RegexPatternString: r.patternString,
			},
		},
	})

	return err
}

func (r *WAFRegionalRegexPatternString) Properties() types.Properties {
	return types.NewProperties().
		Set("RegexPatternSetID", r.patternSetID).
		Set("patternString", r.patternString)
}

func (r *WAFRegionalRegexPatternString) String() string {
	return *r.patternSetID
}
