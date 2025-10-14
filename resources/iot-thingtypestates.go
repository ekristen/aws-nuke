package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/iot" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const IoTThingTypeStateResource = "IoTThingTypeState"

func init() {
	registry.Register(&registry.Registration{
		Name:     IoTThingTypeStateResource,
		Scope:    nuke.Account,
		Resource: &IoTThingTypeState{},
		Lister:   &IoTThingTypeStateLister{},
	})
}

type IoTThingTypeStateLister struct{}

func (l *IoTThingTypeStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &IoTThingTypeState{
				svc:             svc,
				name:            thingType.ThingTypeName,
				deprecated:      thingType.ThingTypeMetadata.Deprecated,
				deprecatedEpoch: thingType.ThingTypeMetadata.DeprecationDate,
			})
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type IoTThingTypeState struct {
	svc             *iot.IoT
	name            *string
	deprecated      *bool
	deprecatedEpoch *time.Time
}

func (f *IoTThingTypeState) Filter() error {
	// ensure we don't inspect time unless its already deprecated
	if ptr.ToBool(f.deprecated) {
		currentTime := time.Now()
		timeDiff := currentTime.Sub(*f.deprecatedEpoch)
		// Must wait for 300 seconds before deleting a ThingType after deprecation
		// Padding 5 seconds to ensure we are beyond any skew
		if timeDiff < 305 {
			return fmt.Errorf("already deprecated")
		}
	}
	return nil
}

func (f *IoTThingTypeState) Remove(_ context.Context) error {
	_, err := f.svc.DeprecateThingType(&iot.DeprecateThingTypeInput{
		ThingTypeName: f.name,
	})

	return err
}

func (f *IoTThingTypeState) String() string {
	return *f.name
}
