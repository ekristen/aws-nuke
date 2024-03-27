package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTThingGroupResource = "IoTThingGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTThingGroupResource,
		Scope:  nuke.Account,
		Lister: &IoTThingGroupLister{},
	})
}

type IoTThingGroupLister struct{}

func (l *IoTThingGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)
	var thingGroups []*iot.GroupNameAndArn

	params := &iot.ListThingGroupsInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListThingGroups(params)
		if err != nil {
			return nil, err
		}

		thingGroups = append(thingGroups, output.ThingGroups...)

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	for _, thingGroup := range thingGroups {
		output, err := svc.DescribeThingGroup(&iot.DescribeThingGroupInput{
			ThingGroupName: thingGroup.GroupName,
		})
		if err != nil {
			return nil, err
		}

		thingGroupType := "static"
		if output.IndexName != nil {
			thingGroupType = "dynamic"
		}

		resources = append(resources, &IoTThingGroup{
			svc:       svc,
			name:      thingGroup.GroupName,
			version:   output.Version,
			groupType: thingGroupType,
		})
	}

	return resources, nil
}

type IoTThingGroup struct {
	svc       *iot.IoT
	name      *string
	version   *int64
	groupType string
}

func (f *IoTThingGroup) Remove(_ context.Context) error {
	if f.groupType == "dynamic" {
		_, err := f.svc.DeleteDynamicThingGroup(&iot.DeleteDynamicThingGroupInput{
			ThingGroupName:  f.name,
			ExpectedVersion: f.version,
		})

		return err
	}

	_, err := f.svc.DeleteThingGroup(&iot.DeleteThingGroupInput{
		ThingGroupName:  f.name,
		ExpectedVersion: f.version,
	})

	return err
}

func (f *IoTThingGroup) String() string {
	return *f.name
}

func (f *IoTThingGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", f.name)
	properties.Set("Type", f.groupType)

	return properties
}
