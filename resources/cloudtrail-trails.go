package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/cloudtrail" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const CloudTrailTrailResource = "CloudTrailTrail"

func init() {
	registry.Register(&registry.Registration{
		Name:     CloudTrailTrailResource,
		Scope:    nuke.Account,
		Resource: &CloudTrailTrail{},
		Lister:   &CloudTrailTrailLister{},
	})
}

type CloudTrailTrailLister struct{}

func (l *CloudTrailTrailLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := cloudtrail.New(opts.Session)

	resp, err := svc.DescribeTrails(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, trail := range resp.TrailList {
		var tags []*cloudtrail.Tag
		tagResp, err := svc.ListTags(
			&cloudtrail.ListTagsInput{
				ResourceIdList: []*string{trail.TrailARN},
			})
		if err != nil {
			logrus.WithError(err).Error("unable to list tags")
		}
		if tagResp != nil && len(tagResp.ResourceTagList) > 0 {
			tags = tagResp.ResourceTagList[0].TagsList
		}

		resources = append(resources, &CloudTrailTrail{
			svc:  svc,
			name: trail.Name,
			tags: tags,
		})
	}

	return resources, nil
}

type CloudTrailTrail struct {
	svc  *cloudtrail.CloudTrail
	name *string
	tags []*cloudtrail.Tag
}

func (trail *CloudTrailTrail) Remove(_ context.Context) error {
	_, err := trail.svc.DeleteTrail(&cloudtrail.DeleteTrailInput{
		Name: trail.name,
	})
	return err
}

func (trail *CloudTrailTrail) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range trail.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("Name", trail.name)
	return properties
}

func (trail *CloudTrailTrail) String() string {
	return *trail.name
}
