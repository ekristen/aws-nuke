package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type CloudFrontCachePolicy struct {
	svc  *cloudfront.CloudFront
	ID   *string
	Name *string
}

const CloudFrontCachePolicyResource = "CloudFrontCachePolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudFrontCachePolicyResource,
		Scope:  nuke.Account,
		Lister: &CloudFrontCachePolicyLister{},
	})
}

type CloudFrontCachePolicyLister struct{}

func (l *CloudFrontCachePolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListCachePoliciesInput{}

	for {
		resp, err := svc.ListCachePolicies(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.CachePolicyList.Items {
			if *item.Type == "custom" { //nolint:goconst
				resources = append(resources, &CloudFrontCachePolicy{
					svc:  svc,
					ID:   item.CachePolicy.Id,
					Name: item.CachePolicy.CachePolicyConfig.Name,
				})
			}
		}

		if resp.CachePolicyList.NextMarker == nil {
			break
		}

		params.Marker = resp.CachePolicyList.NextMarker
	}

	return resources, nil
}

func (f *CloudFrontCachePolicy) Remove(_ context.Context) error {
	resp, err := f.svc.GetCachePolicy(&cloudfront.GetCachePolicyInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteCachePolicy(&cloudfront.DeleteCachePolicyInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontCachePolicy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Name", f.Name)
	return properties
}

func (f *CloudFrontCachePolicy) String() string {
	return *f.Name
}
