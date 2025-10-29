package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3vectors"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3vectorsIndexResource = "S3vectorsIndex"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3vectorsIndexResource,
		Scope:    nuke.Account,
		Resource: &S3vectorsIndex{},
		Lister:   &S3vectorsIndexLister{},
		DependsOn: []string{
			S3vectorsVectorResource,
		},
	})
}

type S3vectorsIndexLister struct{}

func (l *S3vectorsIndexLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3vectors.NewFromConfig(*opts.Config)

	var resources []resource.Resource

	// First, list all vector buckets
	bucketsParams := &s3vectors.ListVectorBucketsInput{}
	bucketsPaginator := s3vectors.NewListVectorBucketsPaginator(svc, bucketsParams)

	for bucketsPaginator.HasMorePages() {
		bucketsPage, err := bucketsPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// For each bucket, list all indexes
		for _, bucket := range bucketsPage.VectorBuckets {
			indexParams := &s3vectors.ListIndexesInput{
				VectorBucketName: bucket.VectorBucketName,
			}

			paginator := s3vectors.NewListIndexesPaginator(svc, indexParams)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, index := range page.Indexes {
					resources = append(resources, &S3vectorsIndex{
						svc:        svc,
						BucketName: bucket.VectorBucketName,
						IndexName:  index.IndexName,
						IndexARN:   index.IndexArn,
					})
				}
			}
		}
	}

	return resources, nil
}

type S3vectorsIndex struct {
	svc        *s3vectors.Client
	BucketName *string
	IndexName  *string
	IndexARN   *string
}

func (r *S3vectorsIndex) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteIndex(ctx, &s3vectors.DeleteIndexInput{
		VectorBucketName: r.BucketName,
		IndexName:        r.IndexName,
	})
	return err
}

func (r *S3vectorsIndex) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3vectorsIndex) String() string {
	return *r.IndexName
}
