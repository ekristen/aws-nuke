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

const WAFv2WebACLResource = "WAFv2WebACL"

func init() {
	registry.Register(&registry.Registration{
		Name:   WAFv2WebACLResource,
		Scope:  nuke.Account,
		Lister: &WAFv2WebACLLister{},
	})
}

type WAFv2WebACLLister struct{}

func (l *WAFv2WebACLLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := wafv2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &wafv2.ListWebACLsInput{
		Limit: aws.Int64(50),
		Scope: aws.String("REGIONAL"),
	}

	output, err := getWebACLs(svc, params)
	if err != nil {
		return []resource.Resource{}, err
	}

	resources = append(resources, output...)

	if *opts.Session.Config.Region == endpoints.UsEast1RegionID {
		params.Scope = aws.String("CLOUDFRONT")

		output, err := getWebACLs(svc, params)
		if err != nil {
			return []resource.Resource{}, err
		}

		resources = append(resources, output...)
	}

	return resources, nil
}

func getWebACLs(svc *wafv2.WAFV2, params *wafv2.ListWebACLsInput) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.ListWebACLs(params)
		if err != nil {
			return nil, err
		}

		for _, webACL := range resp.WebACLs {
			resources = append(resources, &WAFv2WebACL{
				svc:       svc,
				ID:        webACL.Id,
				name:      webACL.Name,
				lockToken: webACL.LockToken,
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

type WAFv2WebACL struct {
	svc       *wafv2.WAFV2
	ID        *string
	name      *string
	lockToken *string
	scope     *string
}

func (f *WAFv2WebACL) Remove(_ context.Context) error {
	_, err := f.svc.DeleteWebACL(&wafv2.DeleteWebACLInput{
		Id:        f.ID,
		Name:      f.name,
		Scope:     f.scope,
		LockToken: f.lockToken,
	})

	return err
}

func (f *WAFv2WebACL) String() string {
	return *f.ID
}

func (f *WAFv2WebACL) Properties() types.Properties {
	properties := types.NewProperties()

	properties.
		Set("ID", f.ID).
		Set("Name", f.name).
		Set("Scope", f.scope)
	return properties
}
