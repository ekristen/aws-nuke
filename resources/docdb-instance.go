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

const DocDBInstanceResource = "DocDBInstance"

func init() {
	registry.Register(&registry.Registration{
		Name:     DocDBInstanceResource,
		Scope:    nuke.Account,
		Resource: &DocDBInstance{},
		Lister:   &DocDBInstanceLister{},
	})
}

type DocDBInstanceLister struct{}

func (l *DocDBInstanceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := docdb.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &docdb.DescribeDBInstancesInput{
		Filters: []docdbtypes.Filter{
			{
				Name:   aws.String("engine"),
				Values: []string{"docdb"},
			},
		},
	}

	paginator := docdb.NewDescribeDBInstancesPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(page.DBInstances); i++ {
			tagList := DocDBEmptyTags
			tags, err := svc.ListTagsForResource(ctx, &docdb.ListTagsForResourceInput{
				ResourceName: page.DBInstances[i].DBInstanceArn,
			})
			if err == nil {
				tagList = tags.TagList
			}
			resources = append(resources, &DocDBInstance{
				svc:        svc,
				Identifier: page.DBInstances[i].DBInstanceIdentifier,
				Tags:       tagList,
			})
		}
	}
	return resources, nil
}

type DocDBInstance struct {
	svc *docdb.Client

	Identifier *string
	Tags       []docdbtypes.Tag
}

func (r *DocDBInstance) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteDBInstance(ctx, &docdb.DeleteDBInstanceInput{
		DBInstanceIdentifier: r.Identifier,
	})
	return err
}

func (r *DocDBInstance) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *DocDBInstance) String() string {
	return *r.Identifier
}
