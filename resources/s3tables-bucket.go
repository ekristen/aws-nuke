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

const S3TablesBucketResource = "S3TablesBucket"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3TablesBucketResource,
		Scope:    nuke.Account,
		Resource: &S3TablesBucket{},
		Lister:   &S3TablesBucketLister{},
		DependsOn: []string{
			S3TablesNamespaceResource,
		},
	})
}

type S3TablesBucketLister struct{}

func (l *S3TablesBucketLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := s3tables.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	paginator := s3tables.NewListTableBucketsPaginator(svc, &s3tables.ListTableBucketsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, tb := range page.TableBuckets {
			tagsResp, err := svc.ListTagsForResource(ctx, &s3tables.ListTagsForResourceInput{
				ResourceArn: tb.Arn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &S3TablesBucket{
				svc:          svc,
				Name:         tb.Name,
				CreationDate: tb.CreatedAt,
				Tags:         tagsResp.Tags,
				Type:         tb.Type,
				arn:          tb.Arn,
			})
		}
	}

	return resources, nil
}

type S3TablesBucket struct {
	svc          *s3tables.Client
	Name         *string                       `description:"The name of the table bucket."`
	CreationDate *time.Time                    `description:"The date and time the table bucket was created."`
	Type         s3tablestypes.TableBucketType `description:"The type of the table bucket (aws or customer)."`
	Tags         map[string]string
	arn          *string
}

func (r *S3TablesBucket) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteTableBucket(ctx, &s3tables.DeleteTableBucketInput{
		TableBucketARN: r.arn,
	})

	return err
}

func (r *S3TablesBucket) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3TablesBucket) String() string {
	return *r.Name
}
