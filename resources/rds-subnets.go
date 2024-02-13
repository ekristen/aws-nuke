package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSDBSubnetGroupResource = "RDSDBSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   RDSDBSubnetGroupResource,
		Scope:  nuke.Account,
		Lister: &RDSDBSubnetGroupLister{},
	})
}

type RDSDBSubnetGroupLister struct{}

type RDSDBSubnetGroup struct {
	svc  *rds.RDS
	name *string
	tags []*rds.Tag
}

func (l *RDSDBSubnetGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBSubnetGroupsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeDBSubnetGroups(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, subnetGroup := range resp.DBSubnetGroups {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: subnetGroup.DBSubnetGroupArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSDBSubnetGroup{
			svc:  svc,
			name: subnetGroup.DBSubnetGroupName,
			tags: tags.TagList,
		})

	}

	return resources, nil
}

func (i *RDSDBSubnetGroup) Remove(_ context.Context) error {
	params := &rds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: i.name,
	}

	_, err := i.svc.DeleteDBSubnetGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSDBSubnetGroup) String() string {
	return *i.name
}

func (i *RDSDBSubnetGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", i.name)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
