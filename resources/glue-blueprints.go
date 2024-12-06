package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueBlueprintResource = "GlueBlueprint"

func init() {
	registry.Register(&registry.Registration{
		Name:     GlueBlueprintResource,
		Scope:    nuke.Account,
		Resource: &GlueBlueprint{},
		Lister:   &GlueBlueprintLister{},
	})
}

type GlueBlueprintLister struct{}

func (l *GlueBlueprintLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.ListBlueprintsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		output, err := svc.ListBlueprints(params)
		if err != nil {
			return nil, err
		}

		for _, blueprint := range output.Blueprints {
			resources = append(resources, &GlueBlueprint{
				svc:  svc,
				name: blueprint,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueBlueprint struct {
	svc  *glue.Glue
	name *string
}

func (f *GlueBlueprint) Remove(_ context.Context) error {
	_, err := f.svc.DeleteBlueprint(&glue.DeleteBlueprintInput{
		Name: f.name,
	})

	return err
}

func (f *GlueBlueprint) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", f.name)

	return properties
}

func (f *GlueBlueprint) String() string {
	return *f.name
}
