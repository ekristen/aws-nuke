package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/xray"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const XRayGroupResource = "XRayGroup"

func init() {
	resource.Register(resource.Registration{
		Name:   XRayGroupResource,
		Scope:  nuke.Account,
		Lister: &XRayGroupLister{},
	})
}

type XRayGroupLister struct{}

func (l *XRayGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := xray.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Get X-Ray Groups
	var xrayGroups []*xray.GroupSummary
	err := svc.GetGroupsPages(
		&xray.GetGroupsInput{},
		func(page *xray.GetGroupsOutput, lastPage bool) bool {
			for _, group := range page.Groups {
				if *group.GroupName != "Default" { // Ignore the Default group as it cannot be removed
					xrayGroups = append(xrayGroups, group)
				}
			}
			return true
		},
	)
	if err != nil {
		return nil, err
	}

	for _, group := range xrayGroups {
		resources = append(resources, &XRayGroup{
			svc:       svc,
			groupName: group.GroupName,
			groupARN:  group.GroupARN,
		})
	}

	return resources, nil
}

type XRayGroup struct {
	svc       *xray.XRay
	groupName *string
	groupARN  *string
}

func (f *XRayGroup) Remove(_ context.Context) error {
	_, err := f.svc.DeleteGroup(&xray.DeleteGroupInput{
		GroupARN: f.groupARN, // Only allowed to pass GroupARN _or_ GroupName to delete request
	})

	return err
}

func (f *XRayGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("GroupName", f.groupName).
		Set("GroupARN", f.groupARN)

	return properties
}
