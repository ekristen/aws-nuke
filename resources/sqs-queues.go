package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const SQSQueueResource = "SQSQueue"

func init() {
	registry.Register(&registry.Registration{
		Name:   SQSQueueResource,
		Scope:  nuke.Account,
		Lister: &SQSQueueLister{},
	})
}

type SQSQueueLister struct{}

func (l *SQSQueueLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := sqs.New(opts.Session)

	params := &sqs.ListQueuesInput{}
	resp, err := svc.ListQueues(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, queue := range resp.QueueUrls {
		resources = append(resources, &SQSQueue{
			svc:      svc,
			queueURL: queue,
		})
	}

	return resources, nil
}

type SQSQueue struct {
	svc      *sqs.SQS
	queueURL *string
}

func (f *SQSQueue) Remove(_ context.Context) error {
	_, err := f.svc.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: f.queueURL,
	})

	return err
}

func (f *SQSQueue) String() string {
	return ptr.ToString(f.queueURL)
}
