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

const WAFv2IPSetResource = "WAFv2IPSet"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFv2IPSetResource,
		Scope:  nuke.Account,
		Lister: &WAFv2IPSetLister{},
	})
}

type WAFv2IPSetLister struct{}

func (l *WAFv2IPSetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &wafv2.ListIPSetsInput{
		Limit: aws.Int64(50),
		Scope: aws.String("REGIONAL"),
	}

	output, err := getIPSets(svc, params)
	if err != nil {
		return []resource.Resource{}, err
	}

	resources = append(resources, output...)

	if *opts.Session.Config.Region == endpoints.UsEast1RegionID {
		params.Scope = aws.String("CLOUDFRONT")

		output, err := getIPSets(svc, params)
		if err != nil {
			return []resource.Resource{}, err
		}

		resources = append(resources, output...)
	}

	return resources, nil
}

func getIPSets(svc *wafv2.WAFV2, params *wafv2.ListIPSetsInput) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.ListIPSets(params)
		if err != nil {
			return nil, err
		}

		for _, set := range resp.IPSets {
			resources = append(resources, &WAFv2IPSet{
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

type WAFv2IPSet struct {
	svc       *wafv2.WAFV2
	id        *string
	name      *string
	lockToken *string
	scope     *string
}

func (r *WAFv2IPSet) Remove(_ context.Context) error {
	_, err := r.svc.DeleteIPSet(&wafv2.DeleteIPSetInput{
		Id:        r.id,
		Name:      r.name,
		Scope:     r.scope,
		LockToken: r.lockToken,
	})

	return err
}

func (r *WAFv2IPSet) Properties() types.Properties {
	return types.NewProperties().
		Set("ID", r.id).
		Set("Name", r.name).
		Set("Scope", r.scope)
}

func (r *WAFv2IPSet) String() string {
	return *r.id
}
