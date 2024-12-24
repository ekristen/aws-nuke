//go:build integration

package resources

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/pkg/awsmod"
)

type readSeekCloser struct{ io.ReadSeeker }

func (readSeekCloser) Close() error { return nil }

type TestS3BucketSuite struct {
	suite.Suite
	bucket *string
	svc    *s3.Client
}

func (suite *TestS3BucketSuite) SetupSuite() {
	var err error

	suite.bucket = ptr.String(fmt.Sprintf("aws-nuke-testing-bucket-%d", time.Now().UnixNano()))

	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		suite.T().Fatalf("failed to create session, %v", err)
	}

	// Create S3 service client
	suite.svc = s3.NewFromConfig(cfg)

	// Create the bucket
	_, err = suite.svc.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: suite.bucket,
		CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint("us-west-2"),
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to create bucket, %v", err)
	}

	// enable versioning
	_, err = suite.svc.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: suite.bucket,
		VersioningConfiguration: &s3types.VersioningConfiguration{
			Status: s3types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to enable versioning, %v", err)
	}

	// Set the object lock configuration to governance mode
	_, err = suite.svc.PutObjectLockConfiguration(ctx, &s3.PutObjectLockConfigurationInput{
		Bucket: suite.bucket,
		ObjectLockConfiguration: &s3types.ObjectLockConfiguration{
			ObjectLockEnabled: s3types.ObjectLockEnabledEnabled,
			Rule: &s3types.ObjectLockRule{
				DefaultRetention: &s3types.DefaultRetention{
					Mode: s3types.ObjectLockRetentionModeGovernance,
					Days: aws.Int32(1),
				},
			},
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to set object lock configuration, %v", err)
	}

	// Create an object in the bucket
	_, err = suite.svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket:            suite.bucket,
		Key:               aws.String("test-object"),
		Body:              readSeekCloser{strings.NewReader("test content")},
		ChecksumAlgorithm: s3types.ChecksumAlgorithmCrc32,
	})
	if err != nil {
		suite.T().Fatalf("failed to create object, %v", err)
	}
}

func (suite *TestS3BucketSuite) TearDownSuite() {
	iterator := newS3DeleteVersionListIterator(suite.svc, &s3.ListObjectVersionsInput{
		Bucket: suite.bucket,
	}, true)
	if err := awsmod.NewBatchDeleteWithClient(suite.svc).Delete(context.TODO(), iterator, bypassGovernanceRetention); err != nil {
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			suite.T().Fatalf("failed to delete objects, %v", err)
		}
	}

	iterator2 := newS3ObjectDeleteListIterator(suite.svc, &s3.ListObjectsV2Input{
		Bucket: suite.bucket,
	}, true)
	if err := awsmod.NewBatchDeleteWithClient(suite.svc).Delete(context.TODO(), iterator2, bypassGovernanceRetention); err != nil {
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			suite.T().Fatalf("failed to delete objects, %v", err)
		}
	}

	_, err := suite.svc.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: suite.bucket,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			suite.T().Fatalf("failed to delete bucket, %v", err)
		}
	}
}

type TestS3BucketObjectLockSuite struct {
	TestS3BucketSuite
}

func (suite *TestS3BucketObjectLockSuite) TestS3BucketObjectLock() {
	// Verify the object lock configuration
	result, err := suite.svc.GetObjectLockConfiguration(context.TODO(), &s3.GetObjectLockConfigurationInput{
		Bucket: suite.bucket,
	})
	if err != nil {
		suite.T().Fatalf("failed to get object lock configuration, %v", err)
	}

	assert.Equal(suite.T(), s3types.ObjectLockEnabledEnabled, result.ObjectLockConfiguration.ObjectLockEnabled)
	assert.Equal(suite.T(), s3types.ObjectLockRetentionModeGovernance, result.ObjectLockConfiguration.Rule.DefaultRetention.Mode)
	assert.Equal(suite.T(), int32(1), *result.ObjectLockConfiguration.Rule.DefaultRetention.Days)
}

func (suite *TestS3BucketObjectLockSuite) TestS3BucketRemove() {
	// Create the S3Bucket object
	bucket := &S3Bucket{
		svc:      suite.svc,
		Name:     suite.bucket,
		settings: &libsettings.Setting{},
	}

	err := bucket.Remove(context.TODO())
	assert.Error(suite.T(), err)
}

type TestS3BucketBypassGovernanceSuite struct {
	TestS3BucketSuite
}

func (suite *TestS3BucketBypassGovernanceSuite) TestS3BucketRemoveWithBypass() {
	// Create the S3Bucket object
	bucket := &S3Bucket{
		svc: suite.svc,
		settings: &libsettings.Setting{
			"BypassGovernanceRetention": true,
		},
		Name:       suite.bucket,
		ObjectLock: s3types.ObjectLockEnabledEnabled,
	}

	err := bucket.Remove(context.TODO())
	assert.Nil(suite.T(), err)
}

func TestS3BucketObjectLock(t *testing.T) {
	suite.Run(t, new(TestS3BucketObjectLockSuite))
}

func TestS3BucketBypassGovernance(t *testing.T) {
	suite.Run(t, new(TestS3BucketBypassGovernanceSuite))
}
