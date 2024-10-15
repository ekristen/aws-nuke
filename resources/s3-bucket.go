package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

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
			"RemoveObjectLegalHold",
		},
	})
}

type S3BucketLister struct{}

func (l *S3BucketLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := s3.NewFromConfig(*opts.Config)

	buckets, err := DescribeS3Buckets(ctx, svc)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, bucket := range buckets {
		newBucket := &S3Bucket{
			svc:          svc,
			name:         aws.ToString(bucket.Name),
			creationDate: aws.ToTime(bucket.CreationDate),
			tags:         make([]s3types.Tag, 0),
		}

		lockCfg, err := svc.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
			Bucket: &newBucket.name,
		})
		if err != nil {
			// check if aws error is NoSuchObjectLockConfiguration
			var aerr smithy.APIError
			if errors.As(err, &aerr) {
				if aerr.ErrorCode() != "ObjectLockConfigurationNotFoundError" {
					logrus.WithError(err).Warn("unknown failure during get object lock configuration")
				}
			}
		}

		if lockCfg != nil && lockCfg.ObjectLockConfiguration != nil {
			newBucket.ObjectLock = lockCfg.ObjectLockConfiguration.ObjectLockEnabled
		}

		tags, err := svc.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			var aerr smithy.APIError
			if errors.As(err, &aerr) {
				if aerr.ErrorCode() == "NoSuchTagSet" {
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

type DescribeS3BucketsAPIClient interface {
	Options() s3.Options
	ListBuckets(context.Context, *s3.ListBucketsInput, ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLocation(context.Context, *s3.GetBucketLocationInput, ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}

func DescribeS3Buckets(ctx context.Context, svc DescribeS3BucketsAPIClient) ([]s3types.Bucket, error) {
	resp, err := svc.ListBuckets(ctx, nil)
	if err != nil {
		return nil, err
	}

	buckets := make([]s3types.Bucket, 0)
	for _, out := range resp.Buckets {
		bucketLocationResponse, err := svc.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: out.Name})
		if err != nil {
			continue
		}

		location := string(bucketLocationResponse.LocationConstraint)
		if location == "" {
			location = "us-east-1"
		}

		region := svc.Options().Region
		if region == "" {
			region = "us-east-1"
		}

		if location == region {
			buckets = append(buckets, out)
		}
	}

	return buckets, nil
}

type S3Bucket struct {
	svc          *s3.Client
	settings     *libsettings.Setting
	name         string
	creationDate time.Time
	tags         []s3types.Tag
	ObjectLock   s3types.ObjectLockEnabled
}

func (r *S3Bucket) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: &r.name,
	})
	if err != nil {
		return err
	}

	_, err = r.svc.PutBucketLogging(ctx, &s3.PutBucketLoggingInput{
		Bucket:              &r.name,
		BucketLoggingStatus: &s3types.BucketLoggingStatus{},
	})
	if err != nil {
		return err
	}

	err = r.RemoveAllLegalHolds(ctx)
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

	_, err = r.svc.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &r.name,
	})

	return err
}

func (r *S3Bucket) RemoveAllLegalHolds(ctx context.Context) error {
	if !r.settings.GetBool("RemoveObjectLegalHold") {
		return nil
	}

	if r.ObjectLock == "" || r.ObjectLock != s3types.ObjectLockEnabledEnabled {
		return nil
	}

	params := &s3.ListObjectsV2Input{
		Bucket: &r.name,
	}

	for {
		res, err := r.svc.ListObjectsV2(ctx, params)
		if err != nil {
			return err
		}

		params.ContinuationToken = res.NextContinuationToken

		for _, obj := range res.Contents {
			_, err := r.svc.PutObjectLegalHold(ctx, &s3.PutObjectLegalHoldInput{
				Bucket:    &r.name,
				Key:       obj.Key,
				LegalHold: &s3types.ObjectLockLegalHold{Status: s3types.ObjectLockLegalHoldStatusOff},
			})
			if err != nil {
				return err
			}
		}

		if res.NextContinuationToken == nil {
			break
		}
	}

	return nil
}

func (r *S3Bucket) RemoveAllVersions(ctx context.Context) error {
	params := &s3.ListObjectVersionsInput{
		Bucket: &r.name,
	}

	var setBypass bool
	var opts []func(input *s3.DeleteObjectsInput)
	if r.ObjectLock == s3types.ObjectLockEnabledEnabled &&
		r.settings.GetBool("BypassGovernanceRetention") {
		setBypass = true
		opts = append(opts, bypassGovernanceRetention)
	}

	iterator := newS3DeleteVersionListIterator(r.svc, params, setBypass)
	return awsmod.NewBatchDeleteWithClient(r.svc).Delete(ctx, iterator, opts...)
}

func (r *S3Bucket) RemoveAllObjects(ctx context.Context) error {
	params := &s3.ListObjectsV2Input{
		Bucket: &r.name,
	}

	var setBypass bool
	var opts []func(input *s3.DeleteObjectsInput)
	if r.ObjectLock == s3types.ObjectLockEnabledEnabled &&
		r.settings.GetBool("BypassGovernanceRetention") {
		setBypass = true
		opts = append(opts, bypassGovernanceRetention)
	}

	iterator := newS3ObjectDeleteListIterator(r.svc, params, setBypass)
	return awsmod.NewBatchDeleteWithClient(r.svc).Delete(ctx, iterator, opts...)
}

func (r *S3Bucket) Settings(settings *libsettings.Setting) {
	r.settings = settings
}

func (r *S3Bucket) Properties() types.Properties {
	properties := types.NewProperties().
		Set("Name", r.name).
		Set("CreationDate", r.creationDate.Format(time.RFC3339)).
		Set("ObjectLock", r.ObjectLock)

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
	Paginator                 *s3.ListObjectVersionsPaginator
	objects                   []s3types.ObjectVersion
	lastNotify                time.Time
	BypassGovernanceRetention *bool
	err                       error
}

func newS3DeleteVersionListIterator(
	svc *s3.Client,
	input *s3.ListObjectVersionsInput,
	bypass bool,
	opts ...func(*s3DeleteVersionListIterator)) awsmod.BatchDeleteIterator {
	iter := &s3DeleteVersionListIterator{
		Bucket:                    input.Bucket,
		Paginator:                 s3.NewListObjectVersionsPaginator(svc, input),
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
		if len(iter.objects) > 0 {
			return true
		}
	}

	if !iter.Paginator.HasMorePages() {
		return false
	}

	page, err := iter.Paginator.NextPage(context.TODO())
	if err != nil {
		iter.err = err
		return false
	}

	iter.objects = page.Versions
	for _, entry := range page.DeleteMarkers {
		iter.objects = append(iter.objects, s3types.ObjectVersion{
			Key:       entry.Key,
			VersionId: entry.VersionId,
		})
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
	return iter.err
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
	Paginator                 *s3.ListObjectsV2Paginator
	objects                   []s3types.Object
	lastNotify                time.Time
	BypassGovernanceRetention bool
	err                       error
}

func newS3ObjectDeleteListIterator(
	svc *s3.Client,
	input *s3.ListObjectsV2Input,
	bypass bool,
	opts ...func(*s3ObjectDeleteListIterator)) awsmod.BatchDeleteIterator {
	iter := &s3ObjectDeleteListIterator{
		Bucket:                    input.Bucket,
		Paginator:                 s3.NewListObjectsV2Paginator(svc, input),
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
		if len(iter.objects) > 0 {
			return true
		}
	}

	if !iter.Paginator.HasMorePages() {
		return false
	}

	page, err := iter.Paginator.NextPage(context.TODO())
	if err != nil {
		iter.err = err
		return false
	}

	iter.objects = page.Contents

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
	return iter.err
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
