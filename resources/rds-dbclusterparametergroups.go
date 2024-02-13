package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSDBClusterParameterGroupResource = "RDSDBClusterParameterGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   RDSDBClusterParameterGroupResource,
		Scope:  nuke.Account,
		Lister: &RDSDBClusterParameterGroupLister{},
	})
}

type RDSDBClusterParameterGroupLister struct{}

func (l *RDSDBClusterParameterGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBClusterParameterGroupsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeDBClusterParameterGroups(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, group := range resp.DBClusterParameterGroups {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: group.DBClusterParameterGroupArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSDBClusterParameterGroup{
			svc:  svc,
			name: group.DBClusterParameterGroupName,
			tags: tags.TagList,
		})

	}

	return resources, nil
}

type RDSDBClusterParameterGroup struct {
	svc  *rds.RDS
	name *string
	tags []*rds.Tag
}

func (i *RDSDBClusterParameterGroup) Filter() error {
	if strings.HasPrefix(*i.name, "default.") {
		return fmt.Errorf("cannot delete default parameter group")
	}
	return nil
}

func (i *RDSDBClusterParameterGroup) Remove(_ context.Context) error {
	params := &rds.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: i.name,
	}

	_, err := i.svc.DeleteDBClusterParameterGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSDBClusterParameterGroup) String() string {
	return *i.name
}

func (i *RDSDBClusterParameterGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", i.name)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
