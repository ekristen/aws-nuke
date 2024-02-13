package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const IoTThingResource = "IoTThing"

func init() {
	registry.Register(&registry.Registration{
		Name:   IoTThingResource,
		Scope:  nuke.Account,
		Lister: &IoTThingLister{},
	})
}

type IoTThingLister struct{}

func (l *IoTThingLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := iot.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &iot.ListThingsInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListThings(params)
		if err != nil {
			return nil, err
		}

		// gather dependent principals
		for _, thing := range output.Things {
			t, err := listIoTThingPrincipals(&IoTThing{
				svc:     svc,
				name:    thing.ThingName,
				version: thing.Version,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, t)
		}
		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

// listIoTThingPrincipals lists the principals attached to a thing (helper function)
func listIoTThingPrincipals(f *IoTThing) (*IoTThing, error) {
	params := &iot.ListThingPrincipalsInput{
		ThingName: f.name,
	}

	output, err := f.svc.ListThingPrincipals(params)
	if err != nil {
		return nil, err
	}

	f.principals = output.Principals
	return f, nil
}

type IoTThing struct {
	svc        *iot.IoT
	name       *string
	version    *int64
	principals []*string
}

func (f *IoTThing) Remove(_ context.Context) error {
	// detach attached principals first
	for _, principal := range f.principals {
		_, err := f.svc.DetachThingPrincipal(&iot.DetachThingPrincipalInput{
			Principal: principal,
			ThingName: f.name,
		})
		if err != nil {
			return err
		}
	}

	_, err := f.svc.DeleteThing(&iot.DeleteThingInput{
		ThingName:       f.name,
		ExpectedVersion: f.version,
	})

	return err
}

func (f *IoTThing) String() string {
	return *f.name
}
