package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ElasticacheSubnetGroupResource = "ElasticacheSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   ElasticacheSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &ElasticacheSubnetGroupLister{},
	})
}

type ElasticacheSubnetGroupLister struct {
	mockSvc elasticacheiface.ElastiCacheAPI
}

func (l *ElasticacheSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc elasticacheiface.ElastiCacheAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = elasticache.New(opts.Session)
	}

	params := &elasticache.DescribeCacheSubnetGroupsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeCacheSubnetGroups(params)
	if err != nil {
		return nil, err
	}

	var resources []resource.Resource
	for _, subnetGroup := range resp.CacheSubnetGroups {
		tags, err := svc.ListTagsForResource(&elasticache.ListTagsForResourceInput{
			ResourceName: subnetGroup.CacheSubnetGroupName,
		})
		if err != nil {
			logrus.WithError(err).Error("unable to retrieve tags")
			continue
		}

		resources = append(resources, &ElasticacheSubnetGroup{
			svc:  svc,
			name: subnetGroup.CacheSubnetGroupName,
			Tags: tags.TagList,
		})
	}

	return resources, nil
}

type ElasticacheSubnetGroup struct {
	svc  elasticacheiface.ElastiCacheAPI
	name *string
	Tags []*elasticache.Tag
}

func (i *ElasticacheSubnetGroup) Filter() error {
	if strings.HasPrefix(*i.name, "default") {
		return fmt.Errorf("cannot delete default subnet group")
	}
	return nil
}

func (i *ElasticacheSubnetGroup) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("Name", i.name)

	for _, tag := range i.Tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
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
