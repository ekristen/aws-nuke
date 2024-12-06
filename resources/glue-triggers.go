package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueTriggerResource = "GlueTrigger"

func init() {
	registry.Register(&registry.Registration{
		Name:     GlueTriggerResource,
		Scope:    nuke.Account,
		Resource: &GlueTrigger{},
		Lister:   &GlueTriggerLister{},
	})
}

type GlueTriggerLister struct{}

func (l *GlueTriggerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetTriggersInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetTriggers(params)
		if err != nil {
			return nil, err
		}

		for _, trigger := range output.Triggers {
			resources = append(resources, &GlueTrigger{
				svc:  svc,
				name: trigger.Name,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueTrigger struct {
	svc  *glue.Glue
	name *string
}

func (f *GlueTrigger) Remove(_ context.Context) error {
	_, err := f.svc.DeleteTrigger(&glue.DeleteTriggerInput{
		Name: f.name,
	})

	return err
}

func (f *GlueTrigger) String() string {
	return *f.name
}
