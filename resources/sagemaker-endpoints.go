package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SageMakerEndpointResource = "SageMakerEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     SageMakerEndpointResource,
		Scope:    nuke.Account,
		Resource: &SageMakerEndpoint{},
		Lister:   &SageMakerEndpointLister{},
	})
}

type SageMakerEndpointLister struct{}

func (l *SageMakerEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListEndpointsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range resp.Endpoints {
			resources = append(resources, &SageMakerEndpoint{
				svc:          svc,
				endpointName: endpoint.EndpointName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerEndpoint struct {
	svc          *sagemaker.SageMaker
	endpointName *string
}

func (f *SageMakerEndpoint) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEndpoint(&sagemaker.DeleteEndpointInput{
		EndpointName: f.endpointName,
	})

	return err
}

func (f *SageMakerEndpoint) String() string {
	return *f.endpointName
}
