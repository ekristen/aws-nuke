package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const BatchComputeEnvironmentStateResource = "BatchComputeEnvironmentState"

func init() {
	registry.Register(&registry.Registration{
		Name:   BatchComputeEnvironmentStateResource,
		Scope:  nuke.Account,
		Lister: &BatchComputeEnvironmentStateLister{},
	})
}

type BatchComputeEnvironmentStateLister struct{}

func (l *BatchComputeEnvironmentStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &BatchComputeEnvironmentState{
				svc:                    svc,
				computeEnvironmentName: computeEnvironment.ComputeEnvironmentName,
				state:                  computeEnvironment.State,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type BatchComputeEnvironmentState struct {
	svc                    *batch.Batch
	computeEnvironmentName *string
	state                  *string
}

func (f *BatchComputeEnvironmentState) Remove(_ context.Context) error {
	_, err := f.svc.UpdateComputeEnvironment(&batch.UpdateComputeEnvironmentInput{
		ComputeEnvironment: f.computeEnvironmentName,
		State:              aws.String("DISABLED"),
	})

	return err
}

func (f *BatchComputeEnvironmentState) String() string {
	return *f.computeEnvironmentName
}

func (f *BatchComputeEnvironmentState) Filter() error {
	if strings.ToLower(*f.state) == "disabled" {
		return fmt.Errorf("already disabled")
	}
	return nil
}
