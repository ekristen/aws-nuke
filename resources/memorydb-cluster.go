package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

type MemoryDBCluster struct {
	svc  *memorydb.MemoryDB
	name *string
	tags []*memorydb.Tag
}

const MemoryDBClusterResource = "MemoryDBCluster"

func init() {
	registry.Register(&registry.Registration{
		Name:   MemoryDBClusterResource,
		Scope:  nuke.Account,
		Lister: &MemoryDBClusterLister{},
	})
}

type MemoryDBClusterLister struct{}

func (l *MemoryDBClusterLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := memorydb.New(opts.Session)
	var resources []resource.Resource

	params := &memorydb.DescribeClustersInput{MaxResults: aws.Int64(100)}

	for {
		resp, err := svc.DescribeClusters(params)
		if err != nil {
			return nil, err
		}

		for _, cluster := range resp.Clusters {
			tags, err := svc.ListTags(&memorydb.ListTagsInput{
				ResourceArn: cluster.ARN,
			})

			if err != nil {
				continue
			}

			resources = append(resources, &MemoryDBCluster{
				svc:  svc,
				name: cluster.Name,
				tags: tags.TagList,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (c *MemoryDBCluster) Remove(_ context.Context) error {
	params := &memorydb.DeleteClusterInput{
		ClusterName: c.name,
	}

	_, err := c.svc.DeleteCluster(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *MemoryDBCluster) String() string {
	return *c.name
}

func (c *MemoryDBCluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", c.name)

	for _, tag := range c.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
