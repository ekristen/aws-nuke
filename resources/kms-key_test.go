//go:build integration

package resources

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go/aws"              //nolint:staticcheck
	"github.com/aws/aws-sdk-go/aws/session"      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/dynamodb" //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/ec2"      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/kms"      //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/sts"      //nolint:staticcheck

	liberrors "github.com/ekristen/libnuke/pkg/errors"
	libsettings "github.com/ekristen/libnuke/pkg/settings"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const (
	kmsTestRegion = "us-east-1"

	// kmsGrantTimeout bounds how long a test will wait for a grant to appear after its resource is created, or to be
	// retired after its resource is deleted.
	kmsGrantTimeout  = 5 * time.Minute
	kmsGrantInterval = 5 * time.Second
)

type TestKMSKeySuite struct {
	suite.Suite

	sess      *session.Session
	kmsSvc    *kms.KMS
	ddbSvc    *dynamodb.DynamoDB
	ec2Svc    *ec2.EC2
	accountID string

	// createdKeys is every key the suite created, scheduled for deletion in TearDownSuite so that a failure part way
	// through a test does not leak billable keys.
	createdKeys []*string
}

func (s *TestKMSKeySuite) SetupSuite() {
	cfg := aws.NewConfig().WithRegion(kmsTestRegion)
	s.sess = session.Must(session.NewSession(cfg))

	s.kmsSvc = kms.New(s.sess)
	s.ddbSvc = dynamodb.New(s.sess)
	s.ec2Svc = ec2.New(s.sess)

	ident, err := sts.New(s.sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	s.Require().NoError(err, "integration tests require working AWS credentials")
	s.accountID = ptr.ToString(ident.Account)

	s.T().Logf("running against account %s in %s", s.accountID, kmsTestRegion)
}

func (s *TestKMSKeySuite) TearDownSuite() {
	// ScheduleKeyDeletion has a seven day minimum window, so these keys linger in PendingDeletion. They are tagged so
	// that they can be identified in the test account.
	for _, keyID := range s.createdKeys {
		if _, err := s.kmsSvc.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
			KeyId:               keyID,
			PendingWindowInDays: aws.Int64(7),
		}); err != nil {
			s.T().Logf("cleanup: unable to schedule deletion of key %s: %v", ptr.ToString(keyID), err)
		}
	}
}

// createKey creates a customer managed key owned by the test account and registers it for cleanup.
func (s *TestKMSKeySuite) createKey(purpose string) *kms.KeyMetadata {
	out, err := s.kmsSvc.CreateKey(&kms.CreateKeyInput{
		Description: aws.String(fmt.Sprintf("aws-nuke integration test: %s", purpose)),
		Policy: aws.String(fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": "arn:aws:iam::%s:root"},
					"Action": "kms:*",
					"Resource": "*"
				}
			]
		}`, s.accountID)),
		Tags: []*kms.Tag{
			{TagKey: aws.String("aws-nuke-integration-test"), TagValue: aws.String(purpose)},
		},
	})
	s.Require().NoError(err)

	s.createdKeys = append(s.createdKeys, out.KeyMetadata.KeyId)

	return out.KeyMetadata
}

// keyState returns the current state of a key.
func (s *TestKMSKeySuite) keyState(keyID *string) string {
	out, err := s.kmsSvc.DescribeKey(&kms.DescribeKeyInput{KeyId: keyID})
	s.Require().NoError(err)

	return ptr.ToString(out.KeyMetadata.KeyState)
}

// awaitGrantees polls until the key has service grants (want true) or has none (want false), returning the grantees
// seen and how long the transition took. This is what verifies that grants really are created and retired alongside
// the resources that use the key, which is the assumption the hold behavior rests on.
func (s *TestKMSKeySuite) awaitGrantees(keyID *string, want bool) ([]string, time.Duration) {
	start := time.Now()

	for {
		grantees, err := serviceGrantees(s.kmsSvc, keyID)
		s.Require().NoError(err)

		if (len(grantees) > 0) == want {
			return grantees, time.Since(start)
		}

		if time.Since(start) > kmsGrantTimeout {
			s.Require().Failf("timed out waiting for grants",
				"key %s: wanted grants present=%v after %s, saw %v",
				ptr.ToString(keyID), want, kmsGrantTimeout, grantees)
		}

		time.Sleep(kmsGrantInterval)
	}
}

// assertHeldThenRemovable drives the real removal path the way the queue does: Remove is retried on each pass, so it
// must hold while the key is in use and succeed on a later pass once the grant has been retired. teardown deletes the
// resource holding the grant.
func (s *TestKMSKeySuite) assertHeldThenRemovable(keyID *string, teardown func()) {
	resource := &KMSKey{svc: s.kmsSvc, ID: keyID}

	// While the encrypted resource exists the key must be held, not scheduled for deletion.
	err := resource.Remove(context.TODO())

	var holdErr liberrors.ErrHoldResource
	s.Require().ErrorAs(err, &holdErr, "expected the key to be held while it is in use")
	s.T().Logf("held as expected: %v", err)

	// The critical assertion: the key was not bricked.
	s.Require().Equal(kms.KeyStateEnabled, s.keyState(keyID),
		"key must remain enabled while a resource still depends on it")

	teardown()

	_, elapsed := s.awaitGrantees(keyID, false)
	s.T().Logf("grant retired %s after the resource was deleted", elapsed.Round(time.Second))

	// Once the grant is gone the same call must now succeed, without needing a second nuke run.
	s.Require().NoError(resource.Remove(context.TODO()))
	s.Require().Equal(kms.KeyStatePendingDeletion, s.keyState(keyID))
}

// TestRemoveSchedulesUnusedKey covers the base case: a key nothing depends on is scheduled for deletion.
func (s *TestKMSKeySuite) TestRemoveSchedulesUnusedKey() {
	key := s.createKey("unused-key")

	resource := &KMSKey{svc: s.kmsSvc, ID: key.KeyId}
	s.Require().NoError(resource.Remove(context.TODO()))

	// ScheduleKeyDeletion does not delete the key, it moves it to PendingDeletion and DescribeKey keeps working.
	s.Require().Equal(kms.KeyStatePendingDeletion, s.keyState(key.KeyId))
}

// TestRemoveHeldByDynamoDB verifies the hold and release cycle against a real encrypted DynamoDB table.
func (s *TestKMSKeySuite) TestRemoveHeldByDynamoDB() {
	key := s.createKey("dynamodb-consumer")
	tableName := fmt.Sprintf("aws-nuke-testing-kms-%d", time.Now().UnixNano())

	_, err := s.ddbSvc.CreateTable(&dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: aws.String(dynamodb.ScalarAttributeTypeS)},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: aws.String(dynamodb.KeyTypeHash)},
		},
		SSESpecification: &dynamodb.SSESpecification{
			Enabled:        aws.Bool(true),
			SSEType:        aws.String(dynamodb.SSETypeKms),
			KMSMasterKeyId: key.Arn,
		},
	})
	s.Require().NoError(err)

	s.Require().NoError(s.ddbSvc.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}))

	grantees, elapsed := s.awaitGrantees(key.KeyId, true)
	s.T().Logf("dynamodb grant appeared after %s: %v", elapsed.Round(time.Second), grantees)

	s.assertHeldThenRemovable(key.KeyId, func() {
		_, err := s.ddbSvc.DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String(tableName)})
		s.Require().NoError(err)

		s.Require().NoError(s.ddbSvc.WaitUntilTableNotExists(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}))
	})
}

// TestRemoveHeldByEBSVolume verifies the same cycle against a second service, so that the service principal match is
// not accidentally specific to one service's principal format.
func (s *TestKMSKeySuite) TestRemoveHeldByEBSVolume() {
	key := s.createKey("ebs-consumer")

	azs, err := s.ec2Svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
	s.Require().NoError(err)
	s.Require().NotEmpty(azs.AvailabilityZones)
	az := azs.AvailabilityZones[0].ZoneName

	vol, err := s.ec2Svc.CreateVolume(&ec2.CreateVolumeInput{
		AvailabilityZone: az,
		Size:             aws.Int64(1),
		VolumeType:       aws.String(ec2.VolumeTypeGp3),
		Encrypted:        aws.Bool(true),
		KmsKeyId:         key.Arn,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeVolume),
				Tags: []*ec2.Tag{
					{Key: aws.String("aws-nuke-integration-test"), Value: aws.String("ebs-consumer")},
				},
			},
		},
	})
	s.Require().NoError(err)

	s.Require().NoError(s.ec2Svc.WaitUntilVolumeAvailable(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{vol.VolumeId},
	}))

	grantees, elapsed := s.awaitGrantees(key.KeyId, true)
	s.T().Logf("ebs grant appeared after %s: %v", elapsed.Round(time.Second), grantees)

	s.assertHeldThenRemovable(key.KeyId, func() {
		_, err := s.ec2Svc.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: vol.VolumeId})
		s.Require().NoError(err)

		s.Require().NoError(s.ec2Svc.WaitUntilVolumeDeleted(&ec2.DescribeVolumesInput{
			VolumeIds: []*string{vol.VolumeId},
		}))
	})
}

// TestDeleteInUseKeysSetting verifies the escape hatch: with the setting enabled an in-use key is deleted anyway.
func (s *TestKMSKeySuite) TestDeleteInUseKeysSetting() {
	key := s.createKey("delete-in-use-setting")
	tableName := fmt.Sprintf("aws-nuke-testing-kms-%d", time.Now().UnixNano())

	_, err := s.ddbSvc.CreateTable(&dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: aws.String(dynamodb.ScalarAttributeTypeS)},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: aws.String(dynamodb.KeyTypeHash)},
		},
		SSESpecification: &dynamodb.SSESpecification{
			Enabled:        aws.Bool(true),
			SSEType:        aws.String(dynamodb.SSETypeKms),
			KMSMasterKeyId: key.Arn,
		},
	})
	s.Require().NoError(err)

	defer func() {
		if _, err := s.ddbSvc.DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String(tableName)}); err != nil {
			s.T().Logf("cleanup: unable to delete table %s: %v", tableName, err)
		}
	}()

	s.Require().NoError(s.ddbSvc.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}))
	s.awaitGrantees(key.KeyId, true)

	setting := libsettings.Setting{"DeleteInUseKeys": true}

	resource := &KMSKey{svc: s.kmsSvc, ID: key.KeyId}
	resource.Settings(&setting)

	s.Require().NoError(resource.Remove(context.TODO()))
	s.Require().Equal(kms.KeyStatePendingDeletion, s.keyState(key.KeyId))
}

// TestGrantsFromIAMPrincipalsAreIgnored verifies that a grant held by an IAM principal rather than a service does not
// hold the key. Such grants are not tied to a resource lifecycle, so counting them would block deletion forever.
func (s *TestKMSKeySuite) TestGrantsFromIAMPrincipalsAreIgnored() {
	key := s.createKey("iam-grant")

	_, err := s.kmsSvc.CreateGrant(&kms.CreateGrantInput{
		KeyId:            key.Arn,
		GranteePrincipal: aws.String(fmt.Sprintf("arn:aws:iam::%s:root", s.accountID)),
		Operations:       []*string{aws.String(kms.GrantOperationDecrypt)},
	})
	s.Require().NoError(err)

	grantees, err := serviceGrantees(s.kmsSvc, key.KeyId)
	s.Require().NoError(err)
	s.Require().Empty(grantees, "an IAM principal grant must not count as in use")

	resource := &KMSKey{svc: s.kmsSvc, ID: key.KeyId}
	s.Require().NoError(resource.Remove(context.TODO()))
	s.Require().Equal(kms.KeyStatePendingDeletion, s.keyState(key.KeyId))
}

// TestServicePrincipalGrantsCannotBeSynthesized records whether CreateGrant accepts a service principal as a grantee.
// It documents why these tests need real encrypted resources rather than hand made grants. If this ever starts
// passing, the fixtures above can be replaced with plain CreateGrant calls.
func (s *TestKMSKeySuite) TestServicePrincipalGrantsCannotBeSynthesized() {
	key := s.createKey("synthetic-service-grant")

	_, err := s.kmsSvc.CreateGrant(&kms.CreateGrantInput{
		KeyId:            key.Arn,
		GranteePrincipal: aws.String(fmt.Sprintf("dynamodb.%s.amazonaws.com", kmsTestRegion)),
		Operations:       []*string{aws.String(kms.GrantOperationDecrypt)},
	})

	// This is a probe rather than a behavioral assertion: neither outcome is a bug in aws-nuke. It exists so that the
	// answer is recorded in the test output instead of being guessed at.
	if err == nil {
		s.T().Log("CreateGrant ACCEPTED a service principal as a grantee; the fixtures above could be simplified")
		return
	}

	s.T().Logf("CreateGrant rejected a service principal, so real encrypted resources are required: %v", err)
}

// TestListerPopulatesInUse exercises the real lister end to end and asserts it reports the in use state of a key that
// a live resource depends on.
func (s *TestKMSKeySuite) TestListerPopulatesInUse() {
	key := s.createKey("lister-in-use")
	tableName := fmt.Sprintf("aws-nuke-testing-kms-%d", time.Now().UnixNano())

	_, err := s.ddbSvc.CreateTable(&dynamodb.CreateTableInput{
		TableName:   aws.String(tableName),
		BillingMode: aws.String(dynamodb.BillingModePayPerRequest),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: aws.String(dynamodb.ScalarAttributeTypeS)},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: aws.String(dynamodb.KeyTypeHash)},
		},
		SSESpecification: &dynamodb.SSESpecification{
			Enabled:        aws.Bool(true),
			SSEType:        aws.String(dynamodb.SSETypeKms),
			KMSMasterKeyId: key.Arn,
		},
	})
	s.Require().NoError(err)

	defer func() {
		if _, err := s.ddbSvc.DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String(tableName)}); err != nil {
			s.T().Logf("cleanup: unable to delete table %s: %v", tableName, err)
		}
	}()

	s.Require().NoError(s.ddbSvc.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}))
	s.awaitGrantees(key.KeyId, true)

	lister := &KMSKeyLister{}
	listed, err := lister.List(context.TODO(), &nuke.ListerOpts{
		Region:    &nuke.Region{Name: kmsTestRegion},
		Session:   s.sess,
		AccountID: ptr.String(s.accountID),
	})
	s.Require().NoError(err)

	var found *KMSKey
	for _, r := range listed {
		if candidate, ok := r.(*KMSKey); ok && ptr.ToString(candidate.ID) == ptr.ToString(key.KeyId) {
			found = candidate
			break
		}
	}

	s.Require().NotNil(found, "lister did not return the key created by this test")
	s.Require().True(ptr.ToBool(found.InUse))
	s.Require().Contains(ptr.ToString(found.InUseBy), "dynamodb")
}

func TestKMSKey(t *testing.T) {
	suite.Run(t, new(TestKMSKeySuite))
}
