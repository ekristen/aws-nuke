package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_sqsiface"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_SQSQueues_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSQS := mock_sqsiface.NewMockSQSAPI(ctrl)

	sqsQueueLister := SQSQueueLister{
		mockSvc: mockSQS,
	}

	mockSQS.EXPECT().ListQueues(gomock.Any()).Return(&sqs.ListQueuesOutput{
		QueueUrls: []*string{
			ptr.String("foobar"),
		},
	}, nil)

	mockSQS.EXPECT().ListQueueTags(gomock.Any()).Return(&sqs.ListQueueTagsOutput{
		Tags: map[string]*string{
			"Name": ptr.String("foobar"),
		},
	}, nil)

	resources, err := sqsQueueLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	sqsQueue := resources[0].(*SQSQueue)
	a.Equal("foobar", sqsQueue.String())
	a.Equal("foobar", sqsQueue.Properties().Get("tag:Name"))
}

func Test_Mock_SQSQueue_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSQS := mock_sqsiface.NewMockSQSAPI(ctrl)

	sqsQueue := SQSQueue{
		svc:      mockSQS,
		queueURL: ptr.String("foobar"),
	}

	mockSQS.EXPECT().DeleteQueue(gomock.Eq(&sqs.DeleteQueueInput{
		QueueUrl: sqsQueue.queueURL,
	})).Return(&sqs.DeleteQueueOutput{}, nil)

	err := sqsQueue.Remove(context.TODO())
	a.Nil(err)
}
