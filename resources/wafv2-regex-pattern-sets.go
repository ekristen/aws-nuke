package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/wafv2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const WAFv2RegexPatternSetResource = "WAFv2RegexPatternSet"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFv2RegexPatternSetResource,
		Scope:  nuke.Account,
		Lister: &WAFv2RegexPatternSetLister{},
	})
}

type WAFv2RegexPatternSetLister struct{}

func (l *WAFv2RegexPatternSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &wafv2.ListRegexPatternSetsInput{
		Limit: aws.Int64(50),
		Scope: aws.String("REGIONAL"),
	}

	output, err := getRegexPatternSets(svc, params)
	if err != nil {
		return []resource.Resource{}, err
	}

	resources = append(resources, output...)

	if *opts.Session.Config.Region == endpoints.UsEast1RegionID {
		params.Scope = aws.String("CLOUDFRONT")

		output, err := getRegexPatternSets(svc, params)
		if err != nil {
			return []resource.Resource{}, err
		}

		resources = append(resources, output...)
	}

	return resources, nil
}

func getRegexPatternSets(svc *wafv2.WAFV2, params *wafv2.ListRegexPatternSetsInput) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.ListRegexPatternSets(params)
		if err != nil {
			return nil, err
		}

		for _, set := range resp.RegexPatternSets {
			resources = append(resources, &WAFv2RegexPatternSet{
				svc:       svc,
				id:        set.Id,
				name:      set.Name,
				lockToken: set.LockToken,
				scope:     params.Scope,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.NextMarker = resp.NextMarker
	}

	return resources, nil
}

type WAFv2RegexPatternSet struct {
	svc       *wafv2.WAFV2
	id        *string
	name      *string
	lockToken *string
	scope     *string
}

func (r *WAFv2RegexPatternSet) Remove(_ context.Context) error {
	_, err := r.svc.DeleteRegexPatternSet(&wafv2.DeleteRegexPatternSetInput{
		Id:        r.id,
		Name:      r.name,
		Scope:     r.scope,
		LockToken: r.lockToken,
	})

	return err
}

func (r *WAFv2RegexPatternSet) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", r.id).
		Set("Name", r.name).
		Set("Scope", r.scope)
}

func (r *WAFv2RegexPatternSet) String() string {
	return *r.id
}
