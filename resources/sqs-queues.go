package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const SQSQueueResource = "SQSQueue"

func init() {
	registry.Register(&registry.Registration{
		Name:   SQSQueueResource,
		Scope:  nuke.Account,
		Lister: &SQSQueueLister{},
	})
}

type SQSQueueLister struct {
	mockSvc sqsiface.SQSAPI
}

func (l *SQSQueueLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc sqsiface.SQSAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = sqs.New(opts.Session)
	}

	params := &sqs.ListQueuesInput{}
	resp, err := svc.ListQueues(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, queue := range resp.QueueUrls {
		var tags map[string]*string
		queueTags, err := svc.ListQueueTags(&sqs.ListQueueTagsInput{
			QueueUrl: queue,
		})
		if err != nil {
			logrus.WithError(err).Error("unable to list queue tags")
		}
		if queueTags != nil {
			tags = queueTags.Tags
		}

		resources = append(resources, &SQSQueue{
			svc:      svc,
			queueURL: queue,
			tags:     tags,
		})
	}

	return resources, nil
}

type SQSQueue struct {
	svc      sqsiface.SQSAPI
	queueURL *string
	tags     map[string]*string
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

func (f *SQSQueue) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("QueueURL", f.queueURL)

	for k, v := range f.tags {
		properties.SetTag(ptr.String(k), v)
	}

	return properties
}
