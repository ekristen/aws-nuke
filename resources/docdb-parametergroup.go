package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	docdbtypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const DocDBParameterGroupResource = "DocDBParameterGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBParameterGroupResource,
		Scope:    nuke.Account,
		Resource: &DocDBParameterGroup{},
		Lister:   &DocDBParameterGroupLister{},
	})
}

type DocDBParameterGroupLister struct{}

func (l *DocDBParameterGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdb.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &docdb.DescribeDBClusterParameterGroupsInput{
		Filters: []docdbtypes.Filter{
			{
				Name:   aws.String("engine"),
				Values: []string{"docdb"},
			},
		},
	}

	paginator := docdb.NewDescribeDBClusterParameterGroupsPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, paramGroup := range page.DBClusterParameterGroups {
			tagList := DocDBEmptyTags
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: paramGroup.DBClusterParameterGroupName,
			})
			if err == nil {
				tagList = tags.TagList
			}
			resources = append(resources, &DocDBParameterGroup{
				svc:  svc,
				Name: paramGroup.DBClusterParameterGroupName,
				Tags: tagList,
			})
		}
	}
	return resources, nil
}

type DocDBParameterGroup struct {
	svc *docdb.Client

	Name *string
	Tags []docdbtypes.Tag
}

func (r *DocDBParameterGroup) Filter() error {
	if strings.HasPrefix(*r.Name, "default.") {
		return fmt.Errorf("default parameter group")
	}
	return nil
}

func (r *DocDBParameterGroup) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDBClusterParameterGroup(ctx, &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: r.Name,
	})
	return err
}

func (r *DocDBParameterGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DocDBParameterGroup) String() string {
	return *r.Name
}
