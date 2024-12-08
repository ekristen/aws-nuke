package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SageMakerModelResource = "SageMakerModel"

func init() {
	registry.Register(&registry.Registration{
		Name:     SageMakerModelResource,
		Scope:    nuke.Account,
		Resource: &SageMakerModel{},
		Lister:   &SageMakerModelLister{},
	})
}

type SageMakerModelLister struct{}

func (l *SageMakerModelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sagemaker.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &sagemaker.ListModelsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListModels(params)
		if err != nil {
			return nil, err
		}

		for _, model := range resp.Models {
			resources = append(resources, &SageMakerModel{
				svc:       svc,
				modelName: model.ModelName,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type SageMakerModel struct {
	svc       *sagemaker.SageMaker
	modelName *string
}

func (f *SageMakerModel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteModel(&sagemaker.DeleteModelInput{
		ModelName: f.modelName,
	})

	return err
}

func (f *SageMakerModel) String() string {
	return *f.modelName
}
