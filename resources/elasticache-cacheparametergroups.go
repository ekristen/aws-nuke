package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"                 //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/elasticache" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ElasticacheCacheParameterGroupResource = "ElasticacheCacheParameterGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     ElasticacheCacheParameterGroupResource,
		Scope:    nuke.Account,
		Resource: &ElasticacheCacheParameterGroup{},
		Lister:   &ElasticacheCacheParameterGroupLister{},
	})
}

type ElasticacheCacheParameterGroupLister struct{}

func (l *ElasticacheCacheParameterGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := elasticache.New(opts.Session)
	var resources []resource.Resource

	params := &elasticache.DescribeCacheParameterGroupsInput{MaxRecords: aws.Int64(100)}

	for {
		resp, err := svc.DescribeCacheParameterGroups(params)
		if err != nil {
			return nil, err
		}

		for _, cacheParameterGroup := range resp.CacheParameterGroups {
			resources = append(resources, &ElasticacheCacheParameterGroup{
				svc:         svc,
				groupName:   cacheParameterGroup.CacheParameterGroupName,
				groupFamily: cacheParameterGroup.CacheParameterGroupFamily,
			})
		}

		if resp.Marker == nil {
			break
		}

		params.Marker = resp.Marker
	}

	return resources, nil
}

type ElasticacheCacheParameterGroup struct {
	svc         *elasticache.ElastiCache
	groupName   *string
	groupFamily *string
}

func (i *ElasticacheCacheParameterGroup) Filter() error {
	if strings.HasPrefix(*i.groupName, "default.") {
		return fmt.Errorf("cannot delete default cache parameter group")
	}
	return nil
}

func (i *ElasticacheCacheParameterGroup) Remove(_ context.Context) error {
	params := &elasticache.DeleteCacheParameterGroupInput{
		CacheParameterGroupName: i.groupName,
	}

	_, err := i.svc.DeleteCacheParameterGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *ElasticacheCacheParameterGroup) String() string {
	return *i.groupName
}

func (i *ElasticacheCacheParameterGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("GroupName", i.groupName).
		Set("GroupFamily", i.groupFamily)
	return properties
}
