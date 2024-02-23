package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

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

type ElasticacheCacheClusterLister struct{}

func (l *ElasticacheCacheClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticache.New(opts.Session)

	params := &elasticache.DescribeCacheClustersInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeCacheClusters(params)
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, cacheCluster := range resp.CacheClusters {
		resources = append(resources, &ElasticacheCacheCluster{
			svc:       svc,
			clusterID: cacheCluster.CacheClusterId,
			status:    cacheCluster.CacheClusterStatus,
		})
	}

	return resources, nil
}

type ElasticacheCacheCluster struct {
	svc       *elasticache.ElastiCache
	clusterID *string
	status    *string
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
