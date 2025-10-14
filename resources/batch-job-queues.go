package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"           //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/batch" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BatchJobQueueResource = "BatchJobQueue"

func init() {
	registry.Register(&registry.Registration{
		Name:     BatchJobQueueResource,
		Scope:    nuke.Account,
		Resource: &BatchJobQueue{},
		Lister:   &BatchJobQueueLister{},
	})
}

type BatchJobQueueLister struct{}

func (l *BatchJobQueueLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
			resources = append(resources, &BatchJobQueue{
				svc:      svc,
				jobQueue: queue.JobQueueName,
			})
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type BatchJobQueue struct {
	svc      *batch.Batch
	jobQueue *string
}

func (f *BatchJobQueue) Remove(_ context.Context) error {
	_, err := f.svc.DeleteJobQueue(&batch.DeleteJobQueueInput{
		JobQueue: f.jobQueue,
	})

	return err
}

func (f *BatchJobQueue) String() string {
	return *f.jobQueue
}
