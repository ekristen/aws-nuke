package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const RDSDBParameterGroupResource = "RDSDBParameterGroup"

func init() {
	resource.Register(&resource.Registration{
		Name:   RDSDBParameterGroupResource,
		Scope:  nuke.Account,
		Lister: &RDSDBParameterGroupLister{},
	})
}

type RDSDBParameterGroupLister struct{}

type RDSDBParameterGroup struct {
	svc  *rds.RDS
	name *string
	tags []*rds.Tag
}

func (l *RDSDBParameterGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeDBParameterGroupsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeDBParameterGroups(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, group := range resp.DBParameterGroups {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: group.DBParameterGroupArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSDBParameterGroup{
			svc:  svc,
			name: group.DBParameterGroupName,
			tags: tags.TagList,
		})

	}

	return resources, nil
}

func (i *RDSDBParameterGroup) Filter() error {
	if strings.HasPrefix(*i.name, "default.") {
		return fmt.Errorf("cannot delete default parameter group")
	}
	return nil
}

func (i *RDSDBParameterGroup) Remove(_ context.Context) error {
	params := &rds.DeleteDBParameterGroupInput{
		DBParameterGroupName: i.name,
	}

	_, err := i.svc.DeleteDBParameterGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSDBParameterGroup) String() string {
	return *i.name
}

func (i *RDSDBParameterGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", i.name)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
