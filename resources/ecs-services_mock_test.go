package resources

import (
	"context"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func Test_ECSService_Properties(t *testing.T) {
	r := &ECSService{
		ServiceARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service"),
		ClusterARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"),
		Tags: map[string]string{
			"Environment": "test",
			"Project":     "aws-nuke",
		},
	}

	properties := r.Properties()

	assert.Equal(t, "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service", properties.Get("ServiceARN"))
	assert.Equal(t, "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster", properties.Get("ClusterARN"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "aws-nuke", properties.Get("tag:Project"))
}

func Test_ECSService_String(t *testing.T) {
	r := &ECSService{
		ServiceARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service"),
		ClusterARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"),
	}

	expected := "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service -> " +
		"arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"
	assert.Equal(t, expected, r.String())
}

func Test_Mock_ECSService_List(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockECSServiceClient)

	clusterArn := "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"
	serviceArn := "arn:aws:ecs:us-east-1:123456789012:service/test-cluster/test-svc"

	mockSvc.On("ListClusters", mock.Anything, &ecs.ListClustersInput{
		MaxResults: ptr.Int32(100),
	}).Return(&ecs.ListClustersOutput{
		ClusterArns: []string{clusterArn},
	}, nil)

	mockSvc.On("ListServices", mock.Anything, &ecs.ListServicesInput{
		Cluster:                ptr.String(clusterArn),
		MaxResults:             ptr.Int32(10),
		ResourceManagementType: ecstypes.ResourceManagementTypeCustomer,
	}).Return(&ecs.ListServicesOutput{
		ServiceArns: []string{serviceArn},
	}, nil)

	mockSvc.On("ListTagsForResource", mock.Anything, &ecs.ListTagsForResourceInput{
		ResourceArn: ptr.String(serviceArn),
	}).Return(&ecs.ListTagsForResourceOutput{
		Tags: []ecstypes.Tag{},
	}, nil)

	lister := &ECSServiceLister{mockSvc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	r := resources[0].(*ECSService)
	a.Equal(serviceArn, *r.ServiceARN)
	a.Equal(clusterArn, *r.ClusterARN)

	mockSvc.AssertExpectations(t)
}
