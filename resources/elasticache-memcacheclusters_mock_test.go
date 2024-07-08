package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_elasticacheiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/aws-nuke/v3/pkg/testsuite"
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
				ARN:                aws.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
				CacheClusterId:     aws.String("foobar"),
				CacheClusterStatus: aws.String("available"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
	}).Return(&elasticache.TagListMessage{}, nil)

	resources, err := cacheClusterLister.List(context.TODO(), &nuke.ListerOpts{})
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
				ARN:            aws.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
				CacheClusterId: aws.String("foobar"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("arn:aws:elasticache:us-west-2:123456789012:cluster:foobar"),
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

	resources, err := cacheClusterLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*ElasticacheCacheCluster)
	a.Len(resource.Tags, 2)
	a.Equal("foobar", resource.String())
	a.Equal("foobar", resource.Properties().Get("tag:Name"))
	a.Equal("test", resource.Properties().Get("tag:aws-nuke"))
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
				ARN:            aws.String("foobar:invalid:arn"),
				CacheClusterId: aws.String("foobar"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("foobar:invalid:arn"),
	}).Return(nil, awserr.New(elasticache.ErrCodeInvalidARNFault, elasticache.ErrCodeInvalidARNFault, nil))

	resources, err := cacheClusterLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 0)

	a.True(called, "expected global hook called and log message to be found")
}
