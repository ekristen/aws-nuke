package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3MultipartUploadResource = "S3MultipartUpload"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3MultipartUploadResource,
		Scope:    nuke.Account,
		Resource: &S3MultipartUpload{},
		Lister:   &S3MultipartUploadLister{},
	})
}

type S3MultipartUploadLister struct{}

func (l *S3MultipartUploadLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3.NewFromConfig(*opts.Config)

	resources := make([]resource.Resource, 0)

	buckets, err := DescribeS3Buckets(ctx, svc, opts)
	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets {
		params := &s3.ListMultipartUploadsInput{
			Bucket: bucket.Name,
		}

		for {
			resp, err := svc.ListMultipartUploads(ctx, params)
			if err != nil {
				return nil, err
			}

			for _, upload := range resp.Uploads {
				if upload.Key == nil || upload.UploadId == nil {
					continue
				}

				resources = append(resources, &S3MultipartUpload{
					svc:      svc,
					Bucket:   bucket.Name,
					Key:      upload.Key,
					UploadID: upload.UploadId,
				})
			}

			if *resp.IsTruncated {
				params.KeyMarker = resp.NextKeyMarker
				continue
			}

			break
		}
	}

	return resources, nil
}

type S3MultipartUpload struct {
	svc      *s3.Client
	Bucket   *string
	Key      *string
	UploadID *string
}

func (r *S3MultipartUpload) Remove(ctx context.Context) error {
	params := &s3.AbortMultipartUploadInput{
		Bucket:   r.Bucket,
		Key:      r.Key,
		UploadId: r.UploadID,
	}

	_, err := r.svc.AbortMultipartUpload(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *S3MultipartUpload) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3MultipartUpload) String() string {
	return fmt.Sprintf("s3://%s/%s#%s", *r.Bucket, *r.Key, *r.UploadID)
}
