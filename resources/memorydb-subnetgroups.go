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

type MemoryDBSubnetGroup struct {
	svc  *memorydb.MemoryDB
	name *string
	tags []*memorydb.Tag
}

const MemoryDBSubnetGroupResource = "MemoryDBSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   MemoryDBSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &MemoryDBSubnetGroupLister{},
	})
}

type MemoryDBSubnetGroupLister struct{}

func (l *MemoryDBSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := memorydb.New(opts.Session)
	var resources []resource.Resource

	params := &memorydb.DescribeSubnetGroupsInput{MaxResults: aws.Int64(100)}

	for {
		resp, err := svc.DescribeSubnetGroups(params)
		if err != nil {
			return nil, err
		}
		for _, subnetGroup := range resp.SubnetGroups {
			tags, err := svc.ListTags(&memorydb.ListTagsInput{
				ResourceArn: subnetGroup.ARN,
			})

			if err != nil {
				continue
			}

			resources = append(resources, &MemoryDBSubnetGroup{
				svc:  svc,
				name: subnetGroup.Name,
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

func (i *MemoryDBSubnetGroup) Remove(_ context.Context) error {
	params := &memorydb.DeleteSubnetGroupInput{
		SubnetGroupName: i.name,
	}

	_, err := i.svc.DeleteSubnetGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *MemoryDBSubnetGroup) String() string {
	return *i.name
}

func (i *MemoryDBSubnetGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("Name", i.name)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
