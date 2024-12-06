package resources

import (
	"context"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ElasticacheCacheClusterResource = "ElasticacheCacheCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:     ElasticacheCacheClusterResource,
		Scope:    nuke.Account,
		Resource: &ElasticacheCacheCluster{},
		Lister:   &ElasticacheCacheClusterLister{},
		DependsOn: []string{
			ElasticacheCacheParameterGroupResource,
			ElasticacheReplicationGroupResource,
			ElasticacheSubnetGroupResource,
		},
	})
}

type ElasticacheCacheClusterLister struct {
	mockSvc elasticacheiface.ElastiCacheAPI
}

func (l *ElasticacheCacheClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc elasticacheiface.ElastiCacheAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = elasticache.New(opts.Session)
	}

	params := &elasticache.DescribeCacheClustersInput{MaxRecords: ptr.Int64(100)}
	resp, err := svc.DescribeCacheClusters(params)
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, cacheCluster := range resp.CacheClusters {
		tags, err := l.getResourceTags(svc, cacheCluster.ARN)
		if err != nil {
			logrus.WithError(err).Error("unable to retrieve tags")
		}

		resources = append(resources, &ElasticacheCacheCluster{
			svc:       svc,
			ClusterID: cacheCluster.CacheClusterId,
			Status:    cacheCluster.CacheClusterStatus,
			Tags:      tags,
		})
	}

	serverlessParams := &elasticache.DescribeServerlessCachesInput{MaxResults: ptr.Int64(100)}
	serverlessResp, serverlessErr := svc.DescribeServerlessCaches(serverlessParams)
	if serverlessErr != nil {
		return nil, serverlessErr
	}

	for _, serverlessCache := range serverlessResp.ServerlessCaches {
		var tags []*elasticache.Tag

		if ptr.ToString(serverlessCache.Status) == "available" ||
			ptr.ToString(serverlessCache.Status) == "modifying" {
			tags, err = l.getResourceTags(svc, serverlessCache.ARN)
			if err != nil {
				logrus.WithError(err).Error("unable to retrieve tags")
			}
		}

		resources = append(resources, &ElasticacheCacheCluster{
			svc:        svc,
			Serverless: true,
			ClusterID:  serverlessCache.ServerlessCacheName,
			Status:     serverlessCache.Status,
			Tags:       tags,
		})
	}

	return resources, nil
}

func (l *ElasticacheCacheClusterLister) getResourceTags(svc elasticacheiface.ElastiCacheAPI, arn *string) ([]*elasticache.Tag, error) {
	tags, err := svc.ListTagsForResource(&elasticache.ListTagsForResourceInput{
		ResourceName: arn,
	})
	if err != nil {
		return []*elasticache.Tag{}, err
	}

	return tags.TagList, nil
}

type ElasticacheCacheCluster struct {
	svc        elasticacheiface.ElastiCacheAPI
	ClusterID  *string
	Status     *string
	Serverless bool
	Tags       []*elasticache.Tag
}

func (r *ElasticacheCacheCluster) Remove(_ context.Context) error {
	if r.Serverless {
		_, err := r.svc.DeleteServerlessCache(&elasticache.DeleteServerlessCacheInput{
			ServerlessCacheName: r.ClusterID,
		})

		return err
	}

	_, err := r.svc.DeleteCacheCluster(&elasticache.DeleteCacheClusterInput{
		CacheClusterId: r.ClusterID,
	})

	return err
}

func (r *ElasticacheCacheCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *ElasticacheCacheCluster) String() string {
	return *r.ClusterID
}
