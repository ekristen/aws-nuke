package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_elasticacheiface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func Test_Mock_ElastiCache_SubnetGroup_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	subnetGroup := ElasticacheSubnetGroup{
		svc:  mockElastiCache,
		name: aws.String("foobar"),
	}

	mockElastiCache.EXPECT().DeleteCacheSubnetGroup(&elasticache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: aws.String("foobar"),
	}).Return(&elasticache.DeleteCacheSubnetGroupOutput{}, nil)

	err := subnetGroup.Remove(context.TODO())
	a.Nil(err)
	a.Equal("foobar", *subnetGroup.name)
}

func Test_Mock_ElastiCache_SubnetGroup_List_NoTags(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	subnetGroupsLister := ElasticacheSubnetGroupLister{
		mockSvc: mockElastiCache,
	}

	mockElastiCache.EXPECT().DescribeCacheSubnetGroups(gomock.Any()).Return(&elasticache.DescribeCacheSubnetGroupsOutput{
		CacheSubnetGroups: []*elasticache.CacheSubnetGroup{
			{
				ARN:                  aws.String("arn:aws:elasticache:us-west-2:123456789012:subnet-group:foobar"),
				CacheSubnetGroupName: aws.String("foobar"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("arn:aws:elasticache:us-west-2:123456789012:subnet-group:foobar"),
	}).Return(&elasticache.TagListMessage{}, nil)

	resources, err := subnetGroupsLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)
	a.Len(resources, 1)

	resource := resources[0].(*ElasticacheSubnetGroup)
	a.Len(resource.Tags, 0)

	a.Equal("foobar", resource.String())
}

func Test_Mock_ElastiCache_SubnetGroup_List_WithTags(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockElastiCache := mock_elasticacheiface.NewMockElastiCacheAPI(ctrl)

	subnetGroupsLister := ElasticacheSubnetGroupLister{
		mockSvc: mockElastiCache,
	}

	mockElastiCache.EXPECT().DescribeCacheSubnetGroups(gomock.Any()).Return(&elasticache.DescribeCacheSubnetGroupsOutput{
		CacheSubnetGroups: []*elasticache.CacheSubnetGroup{
			{
				ARN:                  aws.String("arn:aws:elasticache:us-west-2:123456789012:subnet-group:foobar"),
				CacheSubnetGroupName: aws.String("foobar"),
			},
		},
	}, nil)

	mockElastiCache.EXPECT().ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: aws.String("arn:aws:elasticache:us-west-2:123456789012:subnet-group:foobar"),
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

	resources, err := subnetGroupsLister.List(context.TODO(), &nuke.ListerOpts{})
	a.Nil(err)

	a.Len(resources, 1)

	resource := resources[0].(*ElasticacheSubnetGroup)
	a.Len(resource.Tags, 2)

	a.Equal("foobar", resource.String())
	a.Equal("foobar", resource.Properties().Get("tag:Name"))
	a.Equal("test", resource.Properties().Get("tag:aws-nuke"))
}
