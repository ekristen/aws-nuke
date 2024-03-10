package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudFrontResponseHeadersPolicyResource = "CloudFrontResponseHeadersPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudFrontResponseHeadersPolicyResource,
		Scope:  nuke.Account,
		Lister: &CloudFrontResponseHeadersPolicyLister{},
	})
}

type CloudFrontResponseHeadersPolicyLister struct{}

func (l *CloudFrontResponseHeadersPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListResponseHeadersPoliciesInput{}

	for {
		resp, err := svc.ListResponseHeadersPolicies(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.ResponseHeadersPolicyList.Items {
			resources = append(resources, &CloudFrontResponseHeadersPolicy{
				svc:  svc,
				ID:   item.ResponseHeadersPolicy.Id,
				name: item.ResponseHeadersPolicy.ResponseHeadersPolicyConfig.Name,
			})
		}

		if resp.ResponseHeadersPolicyList.NextMarker == nil {
			break
		}

		params.Marker = resp.ResponseHeadersPolicyList.NextMarker
	}

	return resources, nil
}

type CloudFrontResponseHeadersPolicy struct {
	svc  *cloudfront.CloudFront
	ID   *string
	name *string
}

func (f *CloudFrontResponseHeadersPolicy) Filter() error {
	if strings.HasPrefix(*f.name, "Managed-") {
		return fmt.Errorf("cannot delete default CloudFront Response headers policy")
	}
	return nil
}

func (f *CloudFrontResponseHeadersPolicy) Remove(_ context.Context) error {
	resp, err := f.svc.GetResponseHeadersPolicy(&cloudfront.GetResponseHeadersPolicyInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteResponseHeadersPolicy(&cloudfront.DeleteResponseHeadersPolicyInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontResponseHeadersPolicy) String() string {
	return *f.name
}

func (f *CloudFrontResponseHeadersPolicy) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Name", f.name)
	return properties
}
