package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ecs" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_ecsiface"
)

func Test_Mock_ECSTask_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECS := mock_ecsiface.NewMockECSAPI(ctrl)

	mockECS.EXPECT().ListClusters(&ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
	}).Return(&ecs.ListClustersOutput{
		ClusterArns: []*string{
			aws.String("foobar"),
		},
	}, nil)

	mockECS.EXPECT().ListTasks(&ecs.ListTasksInput{
		Cluster:       aws.String("foobar"),
		MaxResults:    aws.Int64(10),
		DesiredStatus: aws.String("RUNNING"),
	}).Return(&ecs.ListTasksOutput{
		TaskArns: []*string{
			aws.String("arn:aws:ecs:us-west-2:123456789012:task/12345678-1234-1234-1234-123456789012"),
		},
	}, nil)

	mockECS.EXPECT().ListTagsForResource(&ecs.ListTagsForResourceInput{
		ResourceArn: aws.String("arn:aws:ecs:us-west-2:123456789012:task/12345678-1234-1234-1234-123456789012"),
	}).Return(&ecs.ListTagsForResourceOutput{
		Tags: []*ecs.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("foobar"),
			},
		},
	}, nil)

	ecsTaskLister := ECSTaskLister{
		mockSvc: mockECS,
	}

	resources, err := ecsTaskLister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_ECSTask_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECS := mock_ecsiface.NewMockECSAPI(ctrl)

	ecsTask := ECSTask{
		svc:        mockECS,
		taskARN:    ptr.String("arn:aws:ecs:us-west-2:123456789012:task/12345678-1234-1234-1234-123456789012"),
		clusterARN: ptr.String("arn:aws:ecs:us-west-2:123456789012:cluster/12345678-1234-1234-1234-123456789012"),
		tags: []*ecs.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("foobar"),
			},
		},
	}

	a.Equal(*ecsTask.taskARN, ecsTask.Properties().Get("TaskARN"))
	a.Equal("foobar", ecsTask.Properties().Get("tag:Name"))

	mockECS.EXPECT().StopTask(gomock.Eq(&ecs.StopTaskInput{
		Cluster: ecsTask.clusterARN,
		Task:    ecsTask.taskARN,
		Reason:  aws.String("Task stopped via AWS Nuke"),
	})).Return(&ecs.StopTaskOutput{}, nil)

	err := ecsTask.Remove(context.TODO())
	a.Nil(err)
}
