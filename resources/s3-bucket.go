package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libsettings "github.com/ekristen/libnuke/pkg/settings"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsmod"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const S3BucketResource = "S3Bucket"

func init() {
	registry.Register(&registry.Registration{
		Name:   S3BucketResource,
		Scope:  nuke.Account,
		Lister: &S3BucketLister{},
		DependsOn: []string{
			S3ObjectResource,
		},
		AlternativeResource: "AWS::S3::Bucket",
		Settings: []string{
			"BypassGovernanceRetention",
		},
	})
}

type S3BucketLister struct{}

func (l *S3BucketLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3.New(opts.Session)

	buckets, err := DescribeS3Buckets(svc)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, bucket := range buckets {
		newBucket := &S3Bucket{
			svc:          svc,
			name:         aws.StringValue(bucket.Name),
			creationDate: aws.TimeValue(bucket.CreationDate),
			tags:         make([]*s3.Tag, 0),
		}

		tags, err := svc.GetBucketTagging(&s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			var aerr awserr.Error
			if errors.As(err, &aerr) {
				if aerr.Code() == "NoSuchTagSet" {
					resources = append(resources, newBucket)
				}
			}
			continue
		}

		newBucket.tags = tags.TagSet
		resources = append(resources, newBucket)
	}

	return resources, nil
}

func DescribeS3Buckets(svc *s3.S3) ([]s3.Bucket, error) {
	resp, err := svc.ListBuckets(nil)
	if err != nil {
		return nil, err
	}

	buckets := make([]s3.Bucket, 0)
	for _, out := range resp.Buckets {
		bucketLocationResponse, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: out.Name})
		if err != nil {
			continue
		}

		location := ptr.ToString(bucketLocationResponse.LocationConstraint)
		if location == "" {
			location = endpoints.UsEast1RegionID
		}

		region := ptr.ToString(svc.Config.Region)
		if region == "" {
			region = endpoints.UsEast1RegionID
		}

		if location == region {
			buckets = append(buckets, *out)
		}
	}

	return buckets, nil
}

type S3Bucket struct {
	svc          *s3.S3
	settings     *libsettings.Setting
	name         string
	creationDate time.Time
	tags         []*s3.Tag
}

func (r *S3Bucket) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
		Bucket: &r.name,
	})
	if err != nil {
		return err
	}

	_, err = r.svc.PutBucketLogging(&s3.PutBucketLoggingInput{
		Bucket:              &r.name,
		BucketLoggingStatus: &s3.BucketLoggingStatus{},
	})
	if err != nil {
		return err
	}

	err = r.RemoveAllVersions(ctx)
	if err != nil {
		return err
	}

	err = r.RemoveAllObjects(ctx)
	if err != nil {
		return err
	}

	_, err = r.svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &r.name,
	})

	return err
}

func (r *S3Bucket) RemoveAllVersions(ctx context.Context) error {
	params := &s3.ListObjectVersionsInput{
		Bucket: &r.name,
	}

	var opts []func(input *s3.DeleteObjectsInput)
	if r.settings.GetBool("BypassGovernanceRetention") {
		opts = append(opts, bypassGovernanceRetention)
	}

	iterator := newS3DeleteVersionListIterator(r.svc, params, r.settings.GetBool("BypassGovernanceRetention"))
	return awsmod.NewBatchDeleteWithClient(r.svc).Delete(ctx, iterator, opts...)
}

func (r *S3Bucket) RemoveAllObjects(ctx context.Context) error {
	params := &s3.ListObjectsInput{
		Bucket: &r.name,
	}

	var opts []func(input *s3.DeleteObjectsInput)
	if r.settings.GetBool("BypassGovernanceRetention") {
		opts = append(opts, bypassGovernanceRetention)
	}

	iterator := newS3ObjectDeleteListIterator(r.svc, params, r.settings.GetBool("BypassGovernanceRetention"))
	return awsmod.NewBatchDeleteWithClient(r.svc).Delete(ctx, iterator, opts...)
}

func (r *S3Bucket) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

func (r *S3Bucket) Properties() types.Properties {
	properties := types.NewProperties().
		Set("Name", r.name).
		Set("CreationDate", r.creationDate.Format(time.RFC3339))

	for _, tag := range r.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (r *S3Bucket) String() string {
	return fmt.Sprintf("s3://%s", r.name)
}

func bypassGovernanceRetention(input *s3.DeleteObjectsInput) {
	input.BypassGovernanceRetention = ptr.Bool(true)
}

type s3DeleteVersionListIterator struct {
	Bucket                    *string
	Paginator                 request.Pagination
	objects                   []*s3.ObjectVersion
	lastNotify                time.Time
	BypassGovernanceRetention *bool
}

func newS3DeleteVersionListIterator(
	svc s3iface.S3API,
	input *s3.ListObjectVersionsInput,
	bypass bool,
	opts ...func(*s3DeleteVersionListIterator)) awsmod.BatchDeleteIterator {
	iter := &s3DeleteVersionListIterator{
		Bucket: input.Bucket,
		Paginator: request.Pagination{
			NewRequest: func() (*request.Request, error) {
				var inCpy *s3.ListObjectVersionsInput
				if input != nil {
					tmp := *input
					inCpy = &tmp
				}
				req, _ := svc.ListObjectVersionsRequest(inCpy)
				return req, nil
			},
		},
		BypassGovernanceRetention: ptr.Bool(bypass),
	}

	for _, opt := range opts {
		opt(iter)
	}

	return iter
}

// Next will use the S3API client to iterate through a list of objects.
func (iter *s3DeleteVersionListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
	}

	if len(iter.objects) == 0 && iter.Paginator.Next() {
		output := iter.Paginator.Page().(*s3.ListObjectVersionsOutput)
		iter.objects = output.Versions

		for _, entry := range output.DeleteMarkers {
			iter.objects = append(iter.objects, &s3.ObjectVersion{
				Key:       entry.Key,
				VersionId: entry.VersionId,
			})
		}
	}

	if len(iter.objects) > 500 && (iter.lastNotify.IsZero() || time.Since(iter.lastNotify) > 120*time.Second) {
		logrus.Infof(
			"S3Bucket: %s - empty bucket operation in progress, this could take a while, please be patient",
			*iter.Bucket)
		iter.lastNotify = time.Now().UTC()
	}

	return len(iter.objects) > 0
}

// Err will return the last known error from Next.
func (iter *s3DeleteVersionListIterator) Err() error {
	return iter.Paginator.Err()
}

// DeleteObject will return the current object to be deleted.
func (iter *s3DeleteVersionListIterator) DeleteObject() awsmod.BatchDeleteObject {
	return awsmod.BatchDeleteObject{
		Object: &s3.DeleteObjectInput{
			Bucket:                    iter.Bucket,
			Key:                       iter.objects[0].Key,
			VersionId:                 iter.objects[0].VersionId,
			BypassGovernanceRetention: iter.BypassGovernanceRetention,
		},
	}
}

type s3ObjectDeleteListIterator struct {
	Bucket                    *string
	Paginator                 request.Pagination
	objects                   []*s3.Object
	lastNotify                time.Time
	BypassGovernanceRetention bool
}

func newS3ObjectDeleteListIterator(
	svc s3iface.S3API,
	input *s3.ListObjectsInput,
	bypass bool,
	opts ...func(*s3ObjectDeleteListIterator)) awsmod.BatchDeleteIterator {
	iter := &s3ObjectDeleteListIterator{
		Bucket: input.Bucket,
		Paginator: request.Pagination{
			NewRequest: func() (*request.Request, error) {
				var inCpy *s3.ListObjectsInput
				if input != nil {
					tmp := *input
					inCpy = &tmp
				}
				req, _ := svc.ListObjectsRequest(inCpy)
				return req, nil
			},
		},
		BypassGovernanceRetention: bypass,
	}

	for _, opt := range opts {
		opt(iter)
	}
	return iter
}

// Next will use the S3API client to iterate through a list of objects.
func (iter *s3ObjectDeleteListIterator) Next() bool {
	if len(iter.objects) > 0 {
		iter.objects = iter.objects[1:]
	}

	if len(iter.objects) == 0 && iter.Paginator.Next() {
		iter.objects = iter.Paginator.Page().(*s3.ListObjectsOutput).Contents
	}

	if len(iter.objects) > 500 && (iter.lastNotify.IsZero() || time.Since(iter.lastNotify) > 120*time.Second) {
		logrus.Infof(
			"S3Bucket: %s - empty bucket operation in progress, this could take a while, please be patient",
			*iter.Bucket)
		iter.lastNotify = time.Now().UTC()
	}

	return len(iter.objects) > 0
}

// Err will return the last known error from Next.
func (iter *s3ObjectDeleteListIterator) Err() error {
	return iter.Paginator.Err()
}

// DeleteObject will return the current object to be deleted.
func (iter *s3ObjectDeleteListIterator) DeleteObject() awsmod.BatchDeleteObject {
	return awsmod.BatchDeleteObject{
		Object: &s3.DeleteObjectInput{
			Bucket:                    iter.Bucket,
			Key:                       iter.objects[0].Key,
			BypassGovernanceRetention: ptr.Bool(iter.BypassGovernanceRetention),
		},
	}
}
