package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3vectors"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3VectorsVectorResource = "S3VectorsVector"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3VectorsVectorResource,
		Scope:    nuke.Account,
		Resource: &S3VectorsVector{},
		Lister:   &S3VectorsVectorLister{},
	})
}

type S3VectorsVectorLister struct{}

func (l *S3VectorsVectorLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
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

			indexPaginator := s3vectors.NewListIndexesPaginator(svc, indexParams)
			for indexPaginator.HasMorePages() {
				indexPage, err := indexPaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				// For each index, list all vectors
				for _, index := range indexPage.Indexes {
					vectorParams := &s3vectors.ListVectorsInput{
						VectorBucketName: bucket.VectorBucketName,
						IndexName:        index.IndexName,
						ReturnMetadata:   false, // Don't need metadata for deletion
						ReturnData:       false, // Don't need vector data for deletion
					}

					vectorPaginator := s3vectors.NewListVectorsPaginator(svc, vectorParams)
					for vectorPaginator.HasMorePages() {
						vectorPage, err := vectorPaginator.NextPage(ctx)
						if err != nil {
							return nil, err
						}

						for _, vector := range vectorPage.Vectors {
							resources = append(resources, &S3VectorsVector{
								svc:              svc,
								VectorBucketName: bucket.VectorBucketName,
								IndexName:        index.IndexName,
								Key:              vector.Key,
							})
						}
					}
				}
			}
		}
	}

	return resources, nil
}

type S3VectorsVector struct {
	svc              *s3vectors.Client
	VectorBucketName *string
	IndexName        *string
	Key              *string
}

func (r *S3VectorsVector) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteVectors(ctx, &s3vectors.DeleteVectorsInput{
		VectorBucketName: r.VectorBucketName,
		IndexName:        r.IndexName,
		Keys:             []string{*r.Key},
	})
	return err
}

func (r *S3VectorsVector) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3VectorsVector) String() string {
	return *r.Key
}
