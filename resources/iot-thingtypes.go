package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTThingTypeResource = "IoTThingType"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTThingTypeResource,
		Scope:  nuke.Account,
		Lister: &IoTThingTypeLister{},
	})
}

type IoTThingTypeLister struct{}

func (l *IoTThingTypeLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListThingTypesInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListThingTypes(params)
		if err != nil {
			return nil, err
		}

		for _, thingType := range output.ThingTypes {
			resources = append(resources, &IoTThingType{
				svc:  svc,
				name: thingType.ThingTypeName,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type IoTThingType struct {
	svc  *iot.IoT
	name *string
}

func (f *IoTThingType) Remove(_ context.Context) error {
	_, err := f.svc.DeleteThingType(&iot.DeleteThingTypeInput{
		ThingTypeName: f.name,
	})

	return err
}

func (f *IoTThingType) String() string {
	return *f.name
}
