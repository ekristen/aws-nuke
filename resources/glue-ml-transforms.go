package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"          //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/glue" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const GlueMLTransformResource = "GlueMLTransform"

func init() {
	registry.Register(&registry.Registration{
		Name:     GlueMLTransformResource,
		Scope:    nuke.Account,
		Resource: &GlueMLTransform{},
		Lister:   &GlueMLTransformLister{},
	})
}

type GlueMLTransformLister struct{}

func (l *GlueMLTransformLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.ListMLTransformsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.ListMLTransforms(params)
		if err != nil {
			return nil, err
		}

		for _, transformID := range output.TransformIds {
			resources = append(resources, &GlueMLTransform{
				svc: svc,
				id:  transformID,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueMLTransform struct {
	svc *glue.Glue
	id  *string
}

func (f *GlueMLTransform) Remove(_ context.Context) error {
	_, err := f.svc.DeleteMLTransform(&glue.DeleteMLTransformInput{
		TransformId: f.id,
	})

	return err
}

func (f *GlueMLTransform) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", f.id)

	return properties
}

func (f *GlueMLTransform) String() string {
	return *f.id
}
