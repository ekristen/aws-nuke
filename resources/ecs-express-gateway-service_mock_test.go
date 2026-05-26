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

func Test_Mock_ECSExpressGatewayService_List(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockECSServiceClient)

	clusterArn := "arn:aws:ecs:us-east-1:123456789012:cluster/test-cluster"
	serviceArn := "arn:aws:ecs:us-east-1:123456789012:service/test-cluster/test-express-svc"

	mockSvc.On("ListClusters", mock.Anything, &ecs.ListClustersInput{
		MaxResults: ptr.Int32(100),
	}).Return(&ecs.ListClustersOutput{
		ClusterArns: []string{clusterArn},
	}, nil)

	mockSvc.On("ListServices", mock.Anything, &ecs.ListServicesInput{
		Cluster:                ptr.String(clusterArn),
		MaxResults:             ptr.Int32(10),
		ResourceManagementType: ecstypes.ResourceManagementTypeEcs,
	}).Return(&ecs.ListServicesOutput{
		ServiceArns: []string{serviceArn},
	}, nil)

	mockSvc.On("ListTagsForResource", mock.Anything, &ecs.ListTagsForResourceInput{
		ResourceArn: ptr.String(serviceArn),
	}).Return(&ecs.ListTagsForResourceOutput{
		Tags: []ecstypes.Tag{
			{Key: ptr.String("Env"), Value: ptr.String("test")},
		},
	}, nil)

	lister := &ECSExpressGatewayServiceLister{mockSvc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	r := resources[0].(*ECSExpressGatewayService)
	a.Equal(serviceArn, *r.ServiceARN)
	a.Equal("test", r.Tags["Env"])

	mockSvc.AssertExpectations(t)
}

func Test_Mock_ECSExpressGatewayService_Remove(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockECSServiceClient)
	serviceArn := "arn:aws:ecs:us-east-1:123456789012:service/test-cluster/test-express-svc"

	r := &ECSExpressGatewayService{
		svc:        mockSvc,
		ServiceARN: ptr.String(serviceArn),
		Tags:       make(map[string]string),
	}

	mockSvc.On("DeleteExpressGatewayService", mock.Anything, &ecs.DeleteExpressGatewayServiceInput{
		ServiceArn: ptr.String(serviceArn),
	}).Return(&ecs.DeleteExpressGatewayServiceOutput{}, nil)

	err := r.Remove(context.TODO())
	a.Nil(err)

	mockSvc.AssertExpectations(t)
}

func Test_Mock_ECSExpressGatewayService_Properties(t *testing.T) {
	a := assert.New(t)

	r := &ECSExpressGatewayService{
		ServiceARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:service/test-cluster/test-express-svc"),
		Tags: map[string]string{
			"Env": "test",
		},
	}

	props := r.Properties()
	a.Equal("arn:aws:ecs:us-east-1:123456789012:service/test-cluster/test-express-svc", props.Get("ServiceARN"))
	a.Equal("test", props.Get("tag:Env"))
}
