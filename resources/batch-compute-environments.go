package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BatchComputeEnvironmentResource = "BatchComputeEnvironment"

func init() {
	registry.Register(&registry.Registration{
		Name:   BatchComputeEnvironmentResource,
		Scope:  nuke.Account,
		Lister: &BatchComputeEnvironmentLister{},
	})
}

type BatchComputeEnvironmentLister struct{}

func (l *BatchComputeEnvironmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := batch.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &batch.DescribeComputeEnvironmentsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeComputeEnvironments(params)
		if err != nil {
			return nil, err
		}

		for _, computeEnvironment := range output.ComputeEnvironments {
			resources = append(resources, &BatchComputeEnvironment{
				svc:                    svc,
				computeEnvironmentName: computeEnvironment.ComputeEnvironmentName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type BatchComputeEnvironment struct {
	svc                    *batch.Batch
	computeEnvironmentName *string
}

func (f *BatchComputeEnvironment) Remove(_ context.Context) error {
	_, err := f.svc.DeleteComputeEnvironment(&batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: f.computeEnvironmentName,
	})

	return err
}

func (f *BatchComputeEnvironment) String() string {
	return *f.computeEnvironmentName
}
