package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudfront" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFrontOriginAccessControlResource = "CloudFrontOriginAccessControl"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFrontOriginAccessControlResource,
		Scope:    nuke.Account,
		Resource: &CloudFrontOriginAccessControl{},
		Lister:   &CloudFrontOriginAccessControlLister{},
	})
}

type CloudFrontOriginAccessControlLister struct{}

func (l *CloudFrontOriginAccessControlLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListOriginAccessControlsInput{}

	for {
		resp, err := svc.ListOriginAccessControls(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.OriginAccessControlList.Items {
			resources = append(resources, &CloudFrontOriginAccessControl{
				svc: svc,
				ID:  item.Id,
			})
		}

		if !*resp.OriginAccessControlList.IsTruncated {
			break
		}

		params.Marker = resp.OriginAccessControlList.NextMarker
	}

	return resources, nil
}

type CloudFrontOriginAccessControl struct {
	svc *cloudfront.CloudFront
	ID  *string
}

func (f *CloudFrontOriginAccessControl) Remove(_ context.Context) error {
	resp, err := f.svc.GetOriginAccessControl(&cloudfront.GetOriginAccessControlInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteOriginAccessControl(&cloudfront.DeleteOriginAccessControlInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontOriginAccessControl) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	return properties
}

func (f *CloudFrontOriginAccessControl) String() string {
	return *f.ID
}
