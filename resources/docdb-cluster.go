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

const DocDBClusterResource = "DocDBCluster"

var DocDBEmptyTags = []docdbtypes.Tag{}

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBClusterResource,
		Scope:    nuke.Account,
		Resource: &DocDBCluster{},
		Lister:   &DocDBClusterLister{},
	})
}

type DocDBClusterLister struct{}

func (l *DocDBClusterLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdb.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	paginator := docdb.NewDescribeDBClustersPaginator(svc, &docdb.DescribeDBClustersInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(page.DBClusters); i++ {
			tagList := DocDBEmptyTags
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: page.DBClusters[i].DBClusterArn,
			})
			if err == nil {
				tagList = tags.TagList
			}
			resources = append(resources, &DocDBCluster{
				svc:                svc,
				ID:                 page.DBClusters[i].DBClusterIdentifier,
				DeletionProtection: page.DBClusters[i].DeletionProtection,
				Tags:               tagList,
			})
		}
	}
	return resources, nil
}

type DocDBCluster struct {
	svc *docdb.Client

	ID                 *string
	DeletionProtection *bool
	Tags               []docdbtypes.Tag
}

func (r *DocDBCluster) Remove(ctx context.Context) error {
	if *r.DeletionProtection {
		_, err := r.svc.ModifyDBCluster(ctx, &docdb.ModifyDBClusterInput{
			DBClusterIdentifier: r.ID,
			DeletionProtection:  aws.Bool(false),
		})
		if err != nil {
			return err
		}
	}

	_, err := r.svc.DeleteDBCluster(ctx, &docdb.DeleteDBClusterInput{
		DBClusterIdentifier: r.ID,
		SkipFinalSnapshot:   aws.Bool(true),
	})
	return err
}

func (r *DocDBCluster) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
