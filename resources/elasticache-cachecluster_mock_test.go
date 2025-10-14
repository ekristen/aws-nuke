package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"          //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/elasticache" //nolint:staticcheck

	"github.com/ekristen/aws-nuke/v3/mocks/mock_elasticacheiface"
	"github.com/ekristen/aws-nuke/v3/pkg/testsuite"
)

func Test_Mock_ElastiCache_CacheCluster_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	cacheCluster := ElasticacheCacheCluster{
		svc:       mockElastiCache,
		ClusterID: ptr.String("foobar"),
	}

	mockElastiCache.EXPECT().DeleteCacheCluster(&elasticache.DeleteCacheClusterInput{
		CacheClusterId: ptr.String("foobar"),
	}).Return(&elasticache.DeleteCacheClusterOutput{}, nil)

	err := cacheCluster.Remove(context.TODO())
	a.Nil(err)
}

func Test_Mock_ElastiCache_CacheCluster_Remove_Serverless(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	cacheCluster := ElasticacheCacheCluster{
		svc:        mockElastiCache,
		ClusterID:  ptr.String("foobar"),
		Serverless: true,
	}

	mockElastiCache.EXPECT().DeleteServerlessCache(&elasticache.DeleteServerlessCacheInput{
		ServerlessCacheName: ptr.String("foobar"),
	}).Return(&elasticache.DeleteServerlessCacheOutput{}, nil)

	err := cacheCluster.Remove(context.TODO())
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
				ARN:                ptr.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
				CacheClusterId:     ptr.String("foobar"),
				CacheClusterStatus: ptr.String("available"),
			},
		},
	}, nil)
	mockElastiCache.EXPECT().DescribeServerlessCaches(gomock.Any()).Return(&elasticache.DescribeServerlessCachesOutput{
		ServerlessCaches: []*elasticache.ServerlessCache{
			{
				ARN:                 ptr.String("arn:aws:elasticache:us-west-2:123456789012:serverless:foobar"),
				ServerlessCacheName: ptr.String("serverless"),
				Status:              ptr.String("available"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: ptr.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
	}).Return(&elasticache.TagListMessage{}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: ptr.String("arn:aws:elasticache:us-west-2:123456789012:serverless:foobar"),
	}).Return(&elasticache.TagListMessage{}, nil)

	resources, err := cacheClusterLister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	resource := resources[0].(*ElasticacheCacheCluster)
	a.Equal("foobar", resource.String())
	serverlessResource := resources[1].(*ElasticacheCacheCluster)
	a.Equal("serverless", serverlessResource.String())
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
				ARN:            ptr.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
				CacheClusterId: ptr.String("foobar"),
			},
		},
	}, nil)
	mockElastiCache.EXPECT().DescribeServerlessCaches(gomock.Any()).Return(&elasticache.DescribeServerlessCachesOutput{
		ServerlessCaches: []*elasticache.ServerlessCache{
			{
				ARN:                 ptr.String("arn:aws:elasticache:us-west-2:123456789012:serverless:foobar"),
				ServerlessCacheName: ptr.String("serverless"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: ptr.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
	}).Return(&elasticache.TagListMessage{
		TagList: []*elasticache.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("foobar"),
			},
			{
				Key:   ptr.String("aws-nuke"),
				Value: ptr.String("test"),
			},
		},
	}, nil)

	resources, err := cacheClusterLister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	resource := resources[0].(*ElasticacheCacheCluster)
	a.Len(resource.Tags, 2)
	a.Equal("foobar", resource.String())
	a.Equal("foobar", resource.Properties().Get("tag:Name"))
	a.Equal("test", resource.Properties().Get("tag:aws-nuke"))

	serverlessResource := resources[1].(*ElasticacheCacheCluster)
	a.Nil(serverlessResource.Tags)
	a.Equal("serverless", serverlessResource.String())
}

func Test_Mock_ElastiCache_CacheCluster_List_TagsInvalidARN(t *testing.T) {
	called := false

	th := testsuite.NewGlobalHook(t, func(t *testing.T, e *logrus.Entry) {
		if !strings.HasSuffix(e.Caller.Function, "resources.(*ElasticacheCacheClusterLister).List") {
			return
		}

		assert.Equal(t, "unable to retrieve tags", e.Message)

		called = true
	})
	defer th.Cleanup()

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
				ARN:            ptr.String("foobar:invalid:arn"),
				CacheClusterId: ptr.String("foobar"),
			},
		},
	}, nil)
	mockElastiCache.EXPECT().DescribeServerlessCaches(gomock.Any()).Return(&elasticache.DescribeServerlessCachesOutput{
		ServerlessCaches: []*elasticache.ServerlessCache{
			{
				ARN:                 ptr.String("foobar:invalid:arn"),
				ServerlessCacheName: ptr.String("serverless"),
				Status:              ptr.String("available"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: ptr.String("foobar:invalid:arn"),
	}).Return(nil, awserr.New(elasticache.ErrCodeInvalidARNFault, elasticache.ErrCodeInvalidARNFault, nil))

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: ptr.String("foobar:invalid:arn"),
	}).Return(nil, awserr.New(elasticache.ErrCodeInvalidARNFault, elasticache.ErrCodeInvalidARNFault, nil))

	resources, err := cacheClusterLister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 2)

	a.True(called, "expected global hook called and log message to be found")
}
