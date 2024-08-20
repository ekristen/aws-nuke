//go:build integration

package resources

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/pkg/awsmod"
)

type TestS3BucketObjectLockSuite struct {
	suite.Suite
	bucket string
	svc    *s3.S3
}

func (suite *TestS3BucketObjectLockSuite) SetupSuite() {
	var err error

	suite.bucket = fmt.Sprintf("aws-nuke-testing-bucket-%d", time.Now().UnixNano())

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	if err != nil {
		suite.T().Fatalf("failed to create session, %v", err)
	}

	// Create S3 service client
	suite.svc = s3.New(sess)

	// Create the bucket
	_, err = suite.svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(suite.bucket),
	})
	if err != nil {
		suite.T().Fatalf("failed to create bucket, %v", err)
	}

	// enable versioning
	_, err = suite.svc.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: aws.String(suite.bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to enable versioning, %v", err)
	}

	// Set the object lock configuration to governance mode
	_, err = suite.svc.PutObjectLockConfiguration(&s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(suite.bucket),
		ObjectLockConfiguration: &s3.ObjectLockConfiguration{
			ObjectLockEnabled: aws.String("Enabled"),
			Rule: &s3.ObjectLockRule{
				DefaultRetention: &s3.DefaultRetention{
					Mode: aws.String("GOVERNANCE"),
					Days: aws.Int64(1),
				},
			},
		},
	})
	if err != nil {
		suite.T().Fatalf("failed to set object lock configuration, %v", err)
	}

	// Create an object in the bucket
	_, err = suite.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(suite.bucket),
		Key:    aws.String("test-object"),
		Body:   aws.ReadSeekCloser(strings.NewReader("test content")),
	})
	if err != nil {
		suite.T().Fatalf("failed to create object, %v", err)
	}
}

func (suite *TestS3BucketObjectLockSuite) TearDownSuite() {
	iterator := newS3DeleteVersionListIterator(suite.svc, &s3.ListObjectVersionsInput{
		Bucket: &suite.bucket,
	}, true)
	if err := awsmod.NewBatchDeleteWithClient(suite.svc).Delete(context.TODO(), iterator, bypassGovernanceRetention); err != nil {
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			suite.T().Fatalf("failed to delete objects, %v", err)
		}
	}

	iterator2 := newS3ObjectDeleteListIterator(suite.svc, &s3.ListObjectsInput{
		Bucket: &suite.bucket,
	}, true)
	if err := awsmod.NewBatchDeleteWithClient(suite.svc).Delete(context.TODO(), iterator2, bypassGovernanceRetention); err != nil {
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			suite.T().Fatalf("failed to delete objects, %v", err)
		}
	}

	_, err := suite.svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(suite.bucket),
	})
	if err != nil {
		if !strings.Contains(err.Error(), "NoSuchBucket") {
			suite.T().Fatalf("failed to delete bucket, %v", err)
		}
	}
}

func (suite *TestS3BucketObjectLockSuite) TestS3BucketObjectLock() {
	// Verify the object lock configuration
	result, err := suite.svc.GetObjectLockConfiguration(&s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(suite.bucket),
	})
	if err != nil {
		suite.T().Fatalf("failed to get object lock configuration, %v", err)
	}

	assert.Equal(suite.T(), "Enabled", *result.ObjectLockConfiguration.ObjectLockEnabled)
	assert.Equal(suite.T(), "GOVERNANCE", *result.ObjectLockConfiguration.Rule.DefaultRetention.Mode)
	assert.Equal(suite.T(), int64(1), *result.ObjectLockConfiguration.Rule.DefaultRetention.Days)
}

func (suite *TestS3BucketObjectLockSuite) TestS3BucketRemove() {
	// Create the S3Bucket object
	bucket := &S3Bucket{
		svc:      suite.svc,
		name:     suite.bucket,
		settings: &libsettings.Setting{},
	}

	err := bucket.Remove(context.TODO())
	assert.Error(suite.T(), err)
}

func (suite *TestS3BucketObjectLockSuite) TestS3BucketRemoveWithBypass() {
	// Create the S3Bucket object
	bucket := &S3Bucket{
		svc:  suite.svc,
		name: suite.bucket,
		settings: &libsettings.Setting{
			"BypassGovernanceRetention": true,
		},
	}

	err := bucket.Remove(context.TODO())
	assert.Nil(suite.T(), err)
}

func TestS3BucketObjectLock(t *testing.T) {
	suite.Run(t, new(TestS3BucketObjectLockSuite))
}
