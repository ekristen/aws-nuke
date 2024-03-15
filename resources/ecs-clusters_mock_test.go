package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/ekristen/aws-nuke/mocks/mock_ecsiface"

	"github.com/ekristen/aws-nuke/pkg/nuke"
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

	mockECS.EXPECT().DescribeClusters(gomock.Any()).Return(&ecs.DescribeClustersOutput{
		Clusters: []*ecs.Cluster{
			{
				ClusterArn: aws.String("foobar"),
				Tags: []*ecs.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("foobar"),
					},
					{
						Key:   aws.String("aws-nuke"),
						Value: aws.String("test"),
					},
				},
			},
		},
	}, nil)

	resources, err := ecsClusterLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	ecsCluster := resources[0].(*ECSCluster)
	a.Equal("foobar", ecsCluster.String())
	a.Equal("foobar", ecsCluster.Properties().Get("tag:Name"))
	a.Equal("test", ecsCluster.Properties().Get("tag:aws-nuke"))
}

func Test_Mock_ECSCluster_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECS := mock_ecsiface.NewMockECSAPI(ctrl)

	ecsCluster := ECSCluster{
		svc: mockECS,
		ARN: ptr.String("foobar"),
	}

	a.Equal("foobar", ecsCluster.String())

	mockECS.EXPECT().DeleteCluster(gomock.Eq(&ecs.DeleteClusterInput{
		Cluster: aws.String("foobar"),
	})).Return(&ecs.DeleteClusterOutput{}, nil)

	err := ecsCluster.Remove(context.TODO())
	a.Nil(err)
}
