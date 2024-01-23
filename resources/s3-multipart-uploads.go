package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const S3MultipartUploadResource = "S3MultipartUpload"

func init() {
	resource.Register(&resource.Registration{
		Name:   S3MultipartUploadResource,
		Scope:  nuke.Account,
		Lister: &S3MultipartUploadLister{},
	})
}

type S3MultipartUploadLister struct{}

func (l *S3MultipartUploadLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3.New(opts.Session)

	resources := make([]resource.Resource, 0)

	buckets, err := DescribeS3Buckets(svc)
	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets {
		params := &s3.ListMultipartUploadsInput{
			Bucket: bucket.Name,
		}

		for {
			resp, err := svc.ListMultipartUploads(params)
			if err != nil {
				return nil, err
			}

			for _, upload := range resp.Uploads {
				if upload.Key == nil || upload.UploadId == nil {
					continue
				}

				resources = append(resources, &S3MultipartUpload{
					svc:      svc,
					bucket:   aws.StringValue(bucket.Name),
					key:      *upload.Key,
					uploadID: *upload.UploadId,
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
	svc      *s3.S3
	bucket   string
	key      string
	uploadID string
}

func (e *S3MultipartUpload) Remove(_ context.Context) error {
	params := &s3.AbortMultipartUploadInput{
		Bucket:   &e.bucket,
		Key:      &e.key,
		UploadId: &e.uploadID,
	}

	_, err := e.svc.AbortMultipartUpload(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *S3MultipartUpload) Properties() types.Properties {
	return types.NewProperties().
		Set("Bucket", e.bucket).
		Set("Key", e.key).
		Set("UploadID", e.uploadID)
}

func (e *S3MultipartUpload) String() string {
	return fmt.Sprintf("s3://%s/%s#%s", e.bucket, e.key, e.uploadID)
}
