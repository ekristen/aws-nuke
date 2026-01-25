package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudfront" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudFrontPublicKeyResource = "CloudFrontPublicKey"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudFrontPublicKeyResource,
		Scope:    nuke.Account,
		Resource: &CloudFrontPublicKey{},
		Lister:   &CloudFrontPublicKeyLister{},
	})
}

type CloudFrontPublicKeyLister struct{}

func (l *CloudFrontPublicKeyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListPublicKeysInput{}

	for {
		resp, err := svc.ListPublicKeys(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.PublicKeyList.Items {
			resources = append(resources, &CloudFrontPublicKey{
				svc:         svc,
				ID:          item.Id,
				name:        item.Name,
				createdTime: item.CreatedTime,
			})
		}

		if resp.PublicKeyList.NextMarker == nil {
			break
		}

		params.Marker = resp.PublicKeyList.NextMarker
	}

	return resources, nil
}

type CloudFrontPublicKey struct {
	svc         *cloudfront.CloudFront
	ID          *string
	name        *string
	createdTime *time.Time
}

func (f *CloudFrontPublicKey) Remove(_ context.Context) error {
	resp, err := f.svc.GetPublicKey(&cloudfront.GetPublicKeyInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeletePublicKey(&cloudfront.DeletePublicKeyInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontPublicKey) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Name", f.name)
	properties.Set("CreatedTime", f.createdTime.Format(time.RFC3339))
	return properties
}

func (f *CloudFrontPublicKey) String() string {
	return *f.name
}
