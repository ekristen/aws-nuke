package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/machinelearning"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const MachineLearningBranchPredictionResource = "MachineLearningBranchPrediction"

func init() {
	registry.Register(&registry.Registration{
		Name:   MachineLearningBranchPredictionResource,
		Scope:  nuke.Account,
		Lister: &MachineLearningBranchPredictionLister{},
	})
}

type MachineLearningBranchPredictionLister struct{}

func (l *MachineLearningBranchPredictionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := machinelearning.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &machinelearning.DescribeBatchPredictionsInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeBatchPredictions(params)
		if err != nil {
			return nil, err
		}

		for _, result := range output.Results {
			resources = append(resources, &MachineLearningBranchPrediction{
				svc: svc,
				ID:  result.BatchPredictionId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MachineLearningBranchPrediction struct {
	svc *machinelearning.MachineLearning
	ID  *string
}

func (f *MachineLearningBranchPrediction) Remove(_ context.Context) error {
	_, err := f.svc.DeleteBatchPrediction(&machinelearning.DeleteBatchPredictionInput{
		BatchPredictionId: f.ID,
	})

	return err
}

func (f *MachineLearningBranchPrediction) String() string {
	return *f.ID
}
