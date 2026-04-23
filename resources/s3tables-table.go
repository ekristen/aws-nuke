package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	s3tablestypes "github.com/aws/aws-sdk-go-v2/service/s3tables/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3TablesTableResource = "S3TablesTable"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3TablesTableResource,
		Scope:    nuke.Account,
		Resource: &S3TablesTable{},
		Lister:   &S3TablesTableLister{},
	})
}

type S3TablesTableLister struct{}

func (l *S3TablesTableLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := s3tables.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	bucketPaginator := s3tables.NewListTableBucketsPaginator(svc, &s3tables.ListTableBucketsInput{})
	for bucketPaginator.HasMorePages() {
		bucketPage, err := bucketPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, tableBucket := range bucketPage.TableBuckets {
			tablePaginator := s3tables.NewListTablesPaginator(svc, &s3tables.ListTablesInput{
				TableBucketARN: tableBucket.Arn,
			})
			for tablePaginator.HasMorePages() {
				tablePage, err := tablePaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, table := range tablePage.Tables {
					tagsResp, err := svc.ListTagsForResource(ctx, &s3tables.ListTagsForResourceInput{
						ResourceArn: table.TableARN,
					})
					if err != nil {
						return nil, err
					}

					resources = append(resources, &S3TablesTable{
						svc:              svc,
						Name:             table.Name,
						Namespace:        &table.Namespace[0],
						CreationDate:     table.CreatedAt,
						TableBucketName:  tableBucket.Name,
						tableBucketARN:   tableBucket.Arn,
						ManagedByService: table.ManagedByService,
						Type:             tableBucket.Type,
						Tags:             tagsResp.Tags,
					})
				}
			}
		}
	}

	return resources, nil
}

type S3TablesTable struct {
	svc              *s3tables.Client
	Name             *string                       `description:"The name of the table."`
	Namespace        *string                       `description:"The namespace the table belongs to."`
	CreationDate     *time.Time                    `description:"The date and time the table was created."`
	TableBucketName  *string                       `description:"The name of the table bucket the table belongs to."`
	ManagedByService *string                       `description:"The AWS service that manages the table, if applicable."`
	Type             s3tablestypes.TableBucketType `description:"The type of the table bucket (aws or customer)."`
	Tags             map[string]string
	tableBucketARN   *string
}

func (r *S3TablesTable) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteTable(ctx, &s3tables.DeleteTableInput{
		Name:           r.Name,
		Namespace:      r.Namespace,
		TableBucketARN: r.tableBucketARN,
	})

	return err
}

func (r *S3TablesTable) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
