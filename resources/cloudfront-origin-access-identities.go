package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudFrontOriginAccessIdentityResource = "CloudFrontOriginAccessIdentity"

func init() {
	resource.Register(&resource.Registration{
		Name:   CloudFrontOriginAccessIdentityResource,
		Scope:  nuke.Account,
		Lister: &CloudFrontOriginAccessIdentityLister{},
	})
}

type CloudFrontOriginAccessIdentityLister struct{}

func (l *CloudFrontOriginAccessIdentityLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)

	for {
		resp, err := svc.ListCloudFrontOriginAccessIdentities(nil)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.CloudFrontOriginAccessIdentityList.Items {
			resources = append(resources, &CloudFrontOriginAccessIdentity{
				svc: svc,
				ID:  item.Id,
			})
		}
		return resources, nil
	}
}

type CloudFrontOriginAccessIdentity struct {
	svc *cloudfront.CloudFront
	ID  *string
}

func (f *CloudFrontOriginAccessIdentity) Remove(_ context.Context) error {
	resp, err := f.svc.GetCloudFrontOriginAccessIdentity(&cloudfront.GetCloudFrontOriginAccessIdentityInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteCloudFrontOriginAccessIdentity(&cloudfront.DeleteCloudFrontOriginAccessIdentityInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontOriginAccessIdentity) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	return properties
}
