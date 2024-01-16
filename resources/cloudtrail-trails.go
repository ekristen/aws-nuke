package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cloudtrail"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const CloudTrailTrailResource = "CloudTrailTrail"

func init() {
	resource.Register(resource.Registration{
		Name:   CloudTrailTrailResource,
		Scope:  nuke.Account,
		Lister: &CloudTrailTrailLister{},
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
		resources = append(resources, &CloudTrailTrail{
			svc:  svc,
			name: trail.Name,
		})

	}
	return resources, nil
}

type CloudTrailTrail struct {
	svc  *cloudtrail.CloudTrail
	name *string
}

func (trail *CloudTrailTrail) Remove(_ context.Context) error {
	_, err := trail.svc.DeleteTrail(&cloudtrail.DeleteTrailInput{
		Name: trail.name,
	})
	return err
}

func (trail *CloudTrailTrail) Properties() types.Properties {
	return types.NewProperties().
		Set("Name", trail.name)
}

func (trail *CloudTrailTrail) String() string {
	return *trail.name
}
