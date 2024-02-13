package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const ElasticacheSubnetGroupResource = "ElasticacheSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticacheSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &ElasticacheSubnetGroupLister{},
	})
}

type ElasticacheSubnetGroupLister struct{}

func (l *ElasticacheSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticache.New(opts.Session)

	params := &elasticache.DescribeCacheSubnetGroupsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeCacheSubnetGroups(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, subnetGroup := range resp.CacheSubnetGroups {
		resources = append(resources, &ElasticacheSubnetGroup{
			svc:  svc,
			name: subnetGroup.CacheSubnetGroupName,
		})

	}

	return resources, nil
}

type ElasticacheSubnetGroup struct {
	svc  *elasticache.ElastiCache
	name *string
}

func (i *ElasticacheSubnetGroup) Filter() error {
	if strings.HasPrefix(*i.name, "default") {
		return fmt.Errorf("cannot delete default subnet group")
	}
	return nil
}

func (i *ElasticacheSubnetGroup) Remove(_ context.Context) error {
	params := &elasticache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: i.name,
	}

	_, err := i.svc.DeleteCacheSubnetGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *ElasticacheSubnetGroup) String() string {
	return *i.name
}
