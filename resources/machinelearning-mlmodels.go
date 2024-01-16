package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/machinelearning"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MachineLearningMLModelResource = "MachineLearningMLModel"

func init() {
	resource.Register(resource.Registration{
		Name:   MachineLearningMLModelResource,
		Scope:  nuke.Account,
		Lister: &MachineLearningMLModelLister{},
	})
}

type MachineLearningMLModelLister struct{}

func (l *MachineLearningMLModelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := machinelearning.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &machinelearning.DescribeMLModelsInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeMLModels(params)
		if err != nil {
			return nil, err
		}

		for _, result := range output.Results {
			resources = append(resources, &MachineLearningMLModel{
				svc: svc,
				ID:  result.MLModelId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MachineLearningMLModel struct {
	svc *machinelearning.MachineLearning
	ID  *string
}

func (f *MachineLearningMLModel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteMLModel(&machinelearning.DeleteMLModelInput{
		MLModelId: f.ID,
	})

	return err
}

func (f *MachineLearningMLModel) String() string {
	return *f.ID
}
