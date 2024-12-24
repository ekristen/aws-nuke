package resources

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

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
		Name:     S3BucketResource,
		Scope:    nuke.Account,
		Resource: &S3Bucket{},
		Lister:   &S3BucketLister{},
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

	buckets, err := DescribeS3Buckets(ctx, svc, opts)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, bucket := range buckets {
		newBucket := &S3Bucket{
			svc:          svc,
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
			Tags:         make([]s3types.Tag, 0),
		}

		lockCfg, err := svc.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
			Bucket: newBucket.Name,
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

		newBucket.Tags = tags.TagSet
		resources = append(resources, newBucket)
	}

	return resources, nil
}

type DescribeS3BucketsAPIClient interface {
	Options() s3.Options
	ListBuckets(context.Context, *s3.ListBucketsInput, ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLocation(context.Context, *s3.GetBucketLocationInput, ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}

func DescribeS3Buckets(ctx context.Context, svc DescribeS3BucketsAPIClient, opts *nuke.ListerOpts) ([]s3types.Bucket, error) {
	buckets := make([]s3types.Bucket, 0)

	params := &s3.ListBucketsInput{
		BucketRegion: ptr.String(opts.Region.Name),
		MaxBuckets:   ptr.Int32(100),
	}

	for {
		resp, err := svc.ListBuckets(ctx, params)
		if err != nil {
			return nil, err
		}

		buckets = append(buckets, resp.Buckets...)

		if resp.ContinuationToken == nil {
			break
		}

		params.ContinuationToken = resp.ContinuationToken
	}

	return buckets, nil
}

type S3Bucket struct {
	svc          *s3.Client
	settings     *libsettings.Setting
	Name         *string
	CreationDate *time.Time
	Tags         []s3types.Tag
	ObjectLock   s3types.ObjectLockEnabled
}

func (r *S3Bucket) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: r.Name,
	})
	if err != nil {
		return err
	}

	_, err = r.svc.PutBucketLogging(ctx, &s3.PutBucketLoggingInput{
		Bucket:              r.Name,
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
		Bucket: r.Name,
	})

	return err
}

func (r *S3Bucket) RemoveAllLegalHolds(ctx context.Context) error {
	if !r.settings.GetBool("RemoveObjectLegalHold") {
		return nil
	}

	if r.ObjectLock != s3types.ObjectLockEnabledEnabled {
		return nil
	}

	params := &s3.ListObjectsV2Input{
		Bucket: r.Name,
	}

	for {
		res, err := r.svc.ListObjectsV2(ctx, params)
		if err != nil {
			return err
		}

		params.ContinuationToken = res.NextContinuationToken

		for _, obj := range res.Contents {
			_, err := r.svc.PutObjectLegalHold(ctx, &s3.PutObjectLegalHoldInput{
				Bucket:    r.Name,
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
		Bucket: r.Name,
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
		Bucket: r.Name,
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
	return types.NewPropertiesFromStruct(r)
}

func (r *S3Bucket) String() string {
	return fmt.Sprintf("s3://%s", *r.Name)
}
