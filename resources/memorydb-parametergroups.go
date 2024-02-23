package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

type MemoryDBParameterGroup struct {
	svc    *memorydb.MemoryDB
	name   *string
	family *string
	tags   []*memorydb.Tag
}

const MemoryDBParameterGroupResource = "MemoryDBParameterGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   MemoryDBParameterGroupResource,
		Scope:  nuke.Account,
		Lister: &MemoryDBParameterGroupLister{},
	})
}

type MemoryDBParameterGroupLister struct{}

func (l *MemoryDBParameterGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := memorydb.New(opts.Session)
	var resources []resource.Resource

	params := &memorydb.DescribeParameterGroupsInput{MaxResults: aws.Int64(100)}

	for {
		resp, err := svc.DescribeParameterGroups(params)
		if err != nil {
			return nil, err
		}

		for _, parameterGroup := range resp.ParameterGroups {
			tags, err := svc.ListTags(&memorydb.ListTagsInput{
				ResourceArn: parameterGroup.ARN,
			})

			if err != nil {
				continue
			}

			resources = append(resources, &MemoryDBParameterGroup{
				svc:    svc,
				name:   parameterGroup.Name,
				family: parameterGroup.Family,
				tags:   tags.TagList,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

func (i *MemoryDBParameterGroup) Filter() error {
	if strings.HasPrefix(*i.name, "default.") {
		return fmt.Errorf("cannot delete default parameter group")
	}
	return nil
}

func (i *MemoryDBParameterGroup) Remove(_ context.Context) error {
	params := &memorydb.DeleteParameterGroupInput{
		ParameterGroupName: i.name,
	}

	_, err := i.svc.DeleteParameterGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *MemoryDBParameterGroup) String() string {
	return *i.name
}

func (i *MemoryDBParameterGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("Name", i.name).
		Set("Family", i.family)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
