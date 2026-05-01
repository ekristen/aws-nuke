package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3TablesNamespaceResource = "S3TablesNamespace"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3TablesNamespaceResource,
		Scope:    nuke.Account,
		Resource: &S3TablesNamespace{},
		Lister:   &S3TablesNamespaceLister{},
		DependsOn: []string{
			S3TablesTableResource,
		},
	})
}

type S3TablesNamespaceLister struct{}

func (l *S3TablesNamespaceLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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
			namespacePaginator := s3tables.NewListNamespacesPaginator(svc, &s3tables.ListNamespacesInput{
				TableBucketARN: tableBucket.Arn,
			})
			for namespacePaginator.HasMorePages() {
				namespacePage, err := namespacePaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, namespace := range namespacePage.Namespaces {
					resources = append(resources, &S3TablesNamespace{
						svc:             svc,
						Name:            &namespace.Namespace[0],
						CreationDate:    namespace.CreatedAt,
						TableBucketName: tableBucket.Name,
						tableBucketARN:  tableBucket.Arn,
					})
				}
			}
		}
	}

	return resources, nil
}

type S3TablesNamespace struct {
	svc             *s3tables.Client
	Name            *string    `description:"The name of the namespace."`
	CreationDate    *time.Time `description:"The date and time the namespace was created."`
	TableBucketName *string    `description:"The name of the table bucket the namespace belongs to."`
	tableBucketARN  *string
}

func (r *S3TablesNamespace) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteNamespace(ctx, &s3tables.DeleteNamespaceInput{
		Namespace:      r.Name,
		TableBucketARN: r.tableBucketARN,
	})

	return err
}

func (r *S3TablesNamespace) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3TablesNamespace) String() string {
	return *r.Name
}
