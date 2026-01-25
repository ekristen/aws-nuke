package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/docdb"
	docdbtypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"
)

const DocDBSubnetGroupResource = "DocDBSubnetGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBSubnetGroupResource,
		Scope:    nuke.Account,
		Resource: &DocDBSubnetGroup{},
		Lister:   &DocDBSubnetGroupLister{},
	})
}

type DocDBSubnetGroupLister struct{}

func (l *DocDBSubnetGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdb.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := docdb.NewDescribeDBSubnetGroupsPaginator(svc, &docdb.DescribeDBSubnetGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, subnetGroup := range page.DBSubnetGroups {
			tagList := DocDBEmptyTags
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: subnetGroup.DBSubnetGroupArn,
			})
			if err == nil {
				tagList = tags.TagList
			}
			resources = append(resources, &DocDBSubnetGroup{
				svc:  svc,
				Name: subnetGroup.DBSubnetGroupName,
				Tags: tagList,
			})
		}
	}
	return resources, nil
}

type DocDBSubnetGroup struct {
	svc *docdb.Client

	Name *string
	Tags []docdbtypes.Tag
}

func (r *DocDBSubnetGroup) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDBSubnetGroup(ctx, &docdb.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: r.Name,
	})
	return err
}

func (r *DocDBSubnetGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DocDBSubnetGroup) String() string {
	return *r.Name
}
