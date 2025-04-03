package resources

import (
	"context"

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

	paginator := docdb.NewDescribeDBClusterParameterGroupsPaginator(svc, &docdb.DescribeDBClusterParameterGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, paramGroup := range page.DBClusterParameterGroups {
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: paramGroup.DBClusterParameterGroupName,
			})
			if err != nil {
				continue
			}
			resources = append(resources, &DocDBParameterGroup{
				svc:  svc,
				Name: aws.ToString(paramGroup.DBClusterParameterGroupName),
				tags: tags.TagList,
			})
		}
	}
	return resources, nil
}

type DocDBParameterGroup struct {
	svc  *docdb.Client
	Name string
	tags []docdbtypes.Tag
}

func (r *DocDBParameterGroup) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDBClusterParameterGroup(ctx, &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(r.Name),
	})
	return err
}

func (r *DocDBParameterGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.Name)

	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}
