package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

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
					Bucket:       bucket.Name,
					CreationDate: bucket.CreationDate,
					Key:          out.Key,
					VersionID:    out.VersionId,
					IsLatest:     out.IsLatest,
				})
			}

			for _, out := range resp.DeleteMarkers {
				if out.Key == nil {
					continue
				}

				resources = append(resources, &S3Object{
					svc:          svc,
					Bucket:       bucket.Name,
					CreationDate: bucket.CreationDate,
					Key:          out.Key,
					VersionID:    out.VersionId,
					IsLatest:     out.IsLatest,
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
	Bucket       *string
	CreationDate *time.Time
	Key          *string
	VersionID    *string
	IsLatest     *bool
}

func (r *S3Object) Remove(ctx context.Context) error {
	params := &s3.DeleteObjectInput{
		Bucket:    r.Bucket,
		Key:       r.Key,
		VersionId: r.VersionID,
	}

	_, err := r.svc.DeleteObject(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *S3Object) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *S3Object) String() string {
	if r.VersionID != nil && *r.VersionID != "null" && !ptr.ToBool(r.IsLatest) {
		return fmt.Sprintf("s3://%s/%s#%s", *r.Bucket, *r.Key, *r.VersionID)
	}
	return fmt.Sprintf("s3://%s/%s", *r.Bucket, *r.Key)
}
