package resources

import (
	"context"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/gotidy/ptr"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/aws-nuke/mocks/mock_ecsiface"
)

func Test_Mock_ECSCluster_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECS := mock_ecsiface.NewMockECSAPI(ctrl)

	ecsClusterLister := ECSClusterLister{
		mockSvc: mockECS,
	}

	mockECS.EXPECT().ListClusters(gomock.Any()).Return(&ecs.ListClustersOutput{
		ClusterArns: []*string{
			aws.String("foobar"),
		},
	}, nil)

	resources, err := ecsClusterLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)
}

func Test_Mock_ECSCluster_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECS := mock_ecsiface.NewMockECSAPI(ctrl)

	iamUser := ECSCluster{
		svc: mockECS,
		ARN: ptr.String("foobar"),
	}

	mockECS.EXPECT().DeleteCluster(gomock.Eq(&ecs.DeleteClusterInput{
		Cluster: aws.String("foobar"),
	})).Return(&ecs.DeleteClusterOutput{}, nil)

	err := iamUser.Remove(context.TODO())
	a.Nil(err)
}
