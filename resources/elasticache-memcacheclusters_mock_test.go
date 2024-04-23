package resources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/ekristen/aws-nuke/mocks/mock_elasticacheiface"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Mock_ElastiCache_CacheCluster_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	cacheCluster := ElasticacheCacheCluster{
		svc:       mockElastiCache,
		clusterID: aws.String("foobar"),
	}

	mockElastiCache.EXPECT().DeleteCacheCluster(&elasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String("foobar"),
	}).Return(&elasticache.DeleteCacheClusterOutput{}, nil)

	err := cacheCluster.Remove(nil)
	a.Nil(err)
}

func Test_Mock_ElastiCache_CacheCluster_List_NoTags(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	cacheClusterLister := ElasticacheCacheClusterLister{
		mockSvc: mockElastiCache,
	}

	mockElastiCache.EXPECT().DescribeCacheClusters(gomock.Any()).Return(&elasticache.DescribeCacheClustersOutput{
		CacheClusters: []*elasticache.CacheCluster{
			{
				CacheClusterId:     aws.String("foobar"),
				CacheClusterStatus: aws.String("available"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("foobar"),
	}).Return(&elasticache.TagListMessage{}, nil)

	resources, err := cacheClusterLister.List(nil, &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*ElasticacheCacheCluster)
	a.Equal("foobar", resource.String())
}

func Test_Mock_ElastiCache_CacheCluster_List_WithTags(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	cacheClusterLister := ElasticacheCacheClusterLister{
		mockSvc: mockElastiCache,
	}

	mockElastiCache.EXPECT().DescribeCacheClusters(gomock.Any()).Return(&elasticache.DescribeCacheClustersOutput{
		CacheClusters: []*elasticache.CacheCluster{
			{
				CacheClusterId: aws.String("foobar"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("foobar"),
	}).Return(&elasticache.TagListMessage{
		TagList: []*elasticache.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("foobar"),
			},
			{
				Key:   aws.String("aws-nuke"),
				Value: aws.String("test"),
			},
		},
	}, nil)

	resources, err := cacheClusterLister.List(nil, &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*ElasticacheCacheCluster)
	a.Len(resource.Tags, 2)
	a.Equal("foobar", resource.String())
	a.Equal("foobar", resource.Properties().Get("tag:Name"))
	a.Equal("test", resource.Properties().Get("tag:aws-nuke"))

}
