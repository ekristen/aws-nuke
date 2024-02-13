package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const GlueDevEndpointResource = "GlueDevEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:   GlueDevEndpointResource,
		Scope:  nuke.Account,
		Lister: &GlueDevEndpointLister{},
	})
}

type GlueDevEndpointLister struct{}

func (l *GlueDevEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := glue.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &glue.GetDevEndpointsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.GetDevEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, devEndpoint := range output.DevEndpoints {
			resources = append(resources, &GlueDevEndpoint{
				svc:          svc,
				endpointName: devEndpoint.EndpointName,
			})
		}

		// This one API can and does return an empty string
		if output.NextToken == nil || len(*output.NextToken) == 0 {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type GlueDevEndpoint struct {
	svc          *glue.Glue
	endpointName *string
}

func (f *GlueDevEndpoint) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDevEndpoint(&glue.DeleteDevEndpointInput{
		EndpointName: f.endpointName,
	})

	return err
}

func (f *GlueDevEndpoint) String() string {
	return *f.endpointName
}
