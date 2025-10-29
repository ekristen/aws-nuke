package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3vectors"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3vectorsBucketResource = "S3vectorsBucket"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3vectorsBucketResource,
		Scope:    nuke.Account,
		Resource: &S3vectorsBucket{},
		Lister:   &S3vectorsBucketLister{},
		DependsOn: []string{
			S3vectorsIndexResource,
		},
	})
}

type S3vectorsBucketLister struct{}

func (l *S3vectorsBucketLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3vectors.NewFromConfig(*opts.Config)

	var resources []resource.Resource
	params := &s3vectors.ListVectorBucketsInput{}

	paginator := s3vectors.NewListVectorBucketsPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, bucket := range page.VectorBuckets {
			resources = append(resources, &S3vectorsBucket{
				svc:  svc,
				Name: bucket.VectorBucketName,
				ARN:  bucket.VectorBucketArn,
			})
		}
	}

	return resources, nil
}

type S3vectorsBucket struct {
	svc  *s3vectors.Client
	Name *string
	ARN  *string
}

func (r *S3vectorsBucket) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteVectorBucket(ctx, &s3vectors.DeleteVectorBucketInput{
		VectorBucketName: r.Name,
	})
	return err
}

func (r *S3vectorsBucket) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3vectorsBucket) String() string {
	return *r.Name
}
