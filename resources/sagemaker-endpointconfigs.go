package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SageMakerEndpointConfigResource = "SageMakerEndpointConfig"

func init() {
	resource.Register(resource.Registration{
		Name:   SageMakerEndpointConfigResource,
		Scope:  nuke.Account,
		Lister: &SageMakerEndpointConfigLister{},
	})
}

type SageMakerEndpointConfigLister struct{}

func (l *SageMakerEndpointConfigLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListEndpointConfigsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListEndpointConfigs(params)
		if err != nil {
			return nil, err
		}

		for _, endpointConfig := range resp.EndpointConfigs {
			resources = append(resources, &SageMakerEndpointConfig{
				svc:                svc,
				endpointConfigName: endpointConfig.EndpointConfigName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerEndpointConfig struct {
	svc                *sagemaker.SageMaker
	endpointConfigName *string
}

func (f *SageMakerEndpointConfig) Remove(_ context.Context) error {
	_, err := f.svc.DeleteEndpointConfig(&sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: f.endpointConfigName,
	})

	return err
}

func (f *SageMakerEndpointConfig) String() string {
	return *f.endpointConfigName
}
