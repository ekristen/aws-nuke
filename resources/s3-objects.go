package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3ObjectResource = "S3Object"

func init() {
	registry.Register(&registry.Registration{
		Name:     S3ObjectResource,
		Scope:    nuke.Account,
		Resource: &S3Object{},
		Lister:   &S3ObjectLister{},
	})
}

type S3ObjectLister struct{}

func (l *S3ObjectLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3.NewFromConfig(*opts.Config)

	resources := make([]resource.Resource, 0)

	buckets, err := DescribeS3Buckets(ctx, svc, opts)
	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets {
		params := &s3.ListObjectVersionsInput{
			Bucket: bucket.Name,
		}

		for {
			resp, err := svc.ListObjectVersions(ctx, params)
			if err != nil {
				return nil, err
			}

			for _, out := range resp.Versions {
				if out.Key == nil {
					continue
				}

				resources = append(resources, &S3Object{
					svc:          svc,
					bucket:       aws.ToString(bucket.Name),
					creationDate: aws.ToTime(bucket.CreationDate),
					key:          *out.Key,
					versionID:    out.VersionId,
					latest:       ptr.ToBool(out.IsLatest),
				})
			}

			for _, out := range resp.DeleteMarkers {
				if out.Key == nil {
					continue
				}

				resources = append(resources, &S3Object{
					svc:          svc,
					bucket:       aws.ToString(bucket.Name),
					creationDate: aws.ToTime(bucket.CreationDate),
					key:          *out.Key,
					versionID:    out.VersionId,
					latest:       ptr.ToBool(out.IsLatest),
				})
			}

			// make sure to list all with more than 1000 objects
			if *resp.IsTruncated {
				params.KeyMarker = resp.NextKeyMarker
				continue
			}

			break
		}
	}

	return resources, nil
}

type S3Object struct {
	svc          *s3.Client
	bucket       string
	creationDate time.Time
	key          string
	versionID    *string
	latest       bool
}

func (e *S3Object) Remove(ctx context.Context) error {
	params := &s3.DeleteObjectInput{
		Bucket:    &e.bucket,
		Key:       &e.key,
		VersionId: e.versionID,
	}

	_, err := e.svc.DeleteObject(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (e *S3Object) Properties() types.Properties {
	return types.NewProperties().
		Set("Bucket", e.bucket).
		Set("Key", e.key).
		Set("VersionID", e.versionID).
		Set("IsLatest", e.latest).
		Set("CreationDate", e.creationDate)
}

func (e *S3Object) String() string {
	if e.versionID != nil && *e.versionID != "null" && !e.latest {
		return fmt.Sprintf("s3://%s/%s#%s", e.bucket, e.key, *e.versionID)
	}
	return fmt.Sprintf("s3://%s/%s", e.bucket, e.key)
}
