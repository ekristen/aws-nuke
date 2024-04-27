package resources

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ElasticacheCacheClusterResource = "ElasticacheCacheCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticacheCacheClusterResource,
		Scope:  nuke.Account,
		Lister: &ElasticacheCacheClusterLister{},
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

	params := &elasticache.DescribeCacheClustersInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeCacheClusters(params)
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, cacheCluster := range resp.CacheClusters {
		tags, err := svc.ListTagsForResource(&elasticache.ListTagsForResourceInput{
			ResourceName: cacheCluster.CacheClusterId,
		})
		if err != nil {
			logrus.WithError(err).Error("unable to retrieve tags")
			continue
		}

		resources = append(resources, &ElasticacheCacheCluster{
			svc:       svc,
			clusterID: cacheCluster.CacheClusterId,
			status:    cacheCluster.CacheClusterStatus,
			Tags:      tags.TagList,
		})
	}

	return resources, nil
}

type ElasticacheCacheCluster struct {
	svc       elasticacheiface.ElastiCacheAPI
	clusterID *string
	status    *string
	Tags      []*elasticache.Tag
}

func (i *ElasticacheCacheCluster) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ClusterID", i.clusterID)
	properties.Set("Status", i.status)

	for _, tag := range i.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (i *ElasticacheCacheCluster) Remove(_ context.Context) error {
	params := &elasticache.DeleteCacheClusterInput{
		CacheClusterId: i.clusterID,
	}

	_, err := i.svc.DeleteCacheCluster(params)
	if err != nil {
		return err
	}
	return nil
}

func (i *ElasticacheCacheCluster) String() string {
	return *i.clusterID
}
