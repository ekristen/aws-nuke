package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/machinelearning"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const MachineLearningDataSourceResource = "MachineLearningDataSource"

func init() {
	registry.Register(&registry.Registration{
		Name:     MachineLearningDataSourceResource,
		Scope:    nuke.Account,
		Resource: &MachineLearningDataSource{},
		Lister:   &MachineLearningDataSourceLister{},
	})
}

type MachineLearningDataSourceLister struct{}

func (l *MachineLearningDataSourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := machinelearning.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &machinelearning.DescribeDataSourcesInput{
		Limit: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeDataSources(params)
		if err != nil {
			return nil, err
		}

		for _, result := range output.Results {
			resources = append(resources, &MachineLearningDataSource{
				svc: svc,
				ID:  result.DataSourceId,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type MachineLearningDataSource struct {
	svc *machinelearning.MachineLearning
	ID  *string
}

func (f *MachineLearningDataSource) Remove(_ context.Context) error {
	_, err := f.svc.DeleteDataSource(&machinelearning.DeleteDataSourceInput{
		DataSourceId: f.ID,
	})

	return err
}

func (f *MachineLearningDataSource) String() string {
	return *f.ID
}
