package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BatchJobQueueStateResource = "BatchJobQueueState"

func init() {
	registry.Register(&registry.Registration{
		Name:     BatchJobQueueStateResource,
		Scope:    nuke.Account,
		Resource: &BatchJobQueueState{},
		Lister:   &BatchJobQueueStateLister{},
	})
}

type BatchJobQueueStateLister struct{}

func (l *BatchJobQueueStateLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := batch.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &batch.DescribeJobQueuesInput{
		MaxResults: aws.Int64(100),
	}

	for {
		output, err := svc.DescribeJobQueues(params)
		if err != nil {
			return nil, err
		}

		for _, queue := range output.JobQueues {
			resources = append(resources, &BatchJobQueueState{
				svc:      svc,
				jobQueue: queue.JobQueueName,
				state:    queue.State,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type BatchJobQueueState struct {
	svc      *batch.Batch
	jobQueue *string
	state    *string
}

func (f *BatchJobQueueState) Remove(_ context.Context) error {
	_, err := f.svc.UpdateJobQueue(&batch.UpdateJobQueueInput{
		JobQueue: f.jobQueue,
		State:    aws.String("DISABLED"),
	})

	return err
}

func (f *BatchJobQueueState) String() string {
	return *f.jobQueue
}

func (f *BatchJobQueueState) Filter() error {
	if strings.EqualFold(ptr.ToString(f.state), "disabled") {
		return fmt.Errorf("already disabled")
	}
	return nil
}
