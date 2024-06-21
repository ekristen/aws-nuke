package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type CloudFrontKeyGroup struct {
	svc              *cloudfront.CloudFront
	ID               *string
	name             *string
	lastModifiedTime *time.Time
}

const CloudFrontKeyGroupResource = "CloudFrontKeyGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   CloudFrontKeyGroupResource,
		Scope:  nuke.Account,
		Lister: &CloudFrontKeyGroupLister{},
	})
}

type CloudFrontKeyGroupLister struct{}

func (l *CloudFrontKeyGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudfront.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &cloudfront.ListKeyGroupsInput{}

	for {
		resp, err := svc.ListKeyGroups(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.KeyGroupList.Items {
			resources = append(resources, &CloudFrontKeyGroup{
				svc:              svc,
				ID:               item.KeyGroup.Id,
				name:             item.KeyGroup.KeyGroupConfig.Name,
				lastModifiedTime: item.KeyGroup.LastModifiedTime,
			})
		}

		if resp.KeyGroupList.NextMarker == nil {
			break
		}

		params.Marker = resp.KeyGroupList.NextMarker
	}

	return resources, nil
}

func (f *CloudFrontKeyGroup) Remove(_ context.Context) error {
	resp, err := f.svc.GetKeyGroup(&cloudfront.GetKeyGroupInput{
		Id: f.ID,
	})
	if err != nil {
		return err
	}

	_, err = f.svc.DeleteKeyGroup(&cloudfront.DeleteKeyGroupInput{
		Id:      f.ID,
		IfMatch: resp.ETag,
	})

	return err
}

func (f *CloudFrontKeyGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.ID)
	properties.Set("Name", f.name)
	properties.Set("LastModifiedTime", f.lastModifiedTime.Format(time.RFC3339))
	return properties
}
