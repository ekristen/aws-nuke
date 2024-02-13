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

const RDSOptionGroupResource = "RDSOptionGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:   RDSOptionGroupResource,
		Scope:  nuke.Account,
		Lister: &RDSOptionGroupLister{},
	})
}

type RDSOptionGroupLister struct{}

func (l *RDSOptionGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := rds.New(opts.Session)

	params := &rds.DescribeOptionGroupsInput{MaxRecords: aws.Int64(100)}
	resp, err := svc.DescribeOptionGroups(params)
	if err != nil {
		return nil, err
	}
	var resources []resource.Resource
	for _, optionGroup := range resp.OptionGroupsList {
		tags, err := svc.ListTagsForResource(&rds.ListTagsForResourceInput{
			ResourceName: optionGroup.OptionGroupArn,
		})

		if err != nil {
			continue
		}

		resources = append(resources, &RDSOptionGroup{
			svc:  svc,
			name: optionGroup.OptionGroupName,
			tags: tags.TagList,
		})

	}

	return resources, nil
}

type RDSOptionGroup struct {
	svc  *rds.RDS
	name *string
	tags []*rds.Tag
}

func (i *RDSOptionGroup) Filter() error {
	if strings.HasPrefix(*i.name, "default:") {
		return fmt.Errorf("cannot delete default Option group")
	}
	return nil
}

func (i *RDSOptionGroup) Remove(_ context.Context) error {
	params := &rds.DeleteOptionGroupInput{
		OptionGroupName: i.name,
	}

	_, err := i.svc.DeleteOptionGroup(params)
	if err != nil {
		return err
	}

	return nil
}

func (i *RDSOptionGroup) String() string {
	return *i.name
}

func (i *RDSOptionGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", i.name)

	for _, tag := range i.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
