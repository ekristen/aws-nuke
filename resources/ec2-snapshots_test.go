package resources

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2Snapshot_String(t *testing.T) {
	a := assert.New(t)

	ec2Snapshot := EC2Snapshot{
		SnapshotID: ptr.String("snap-1234567890abcdef0"),
	}

	a.Equal("snap-1234567890abcdef0", ec2Snapshot.String())
}

func Test_EC2Snapshot_Properties(t *testing.T) {
	a := assert.New(t)

	startTime := time.Now()
	restoreExpiryTime := time.Now().Add(24 * time.Hour)
	state := ec2types.SnapshotStateCompleted
	storageTier := ec2types.StorageTierStandard

	ec2Snapshot := EC2Snapshot{
		SnapshotID:          ptr.String("snap-1234567890abcdef0"),
		Description:         ptr.String("My snapshot"),
		VolumeID:            ptr.String("vol-1234567890abcdef0"),
		VolumeSize:          ptr.Int32(100),
		State:               &state,
		StateMessage:        ptr.String("100% complete"),
		StartTime:           &startTime,
		Progress:            ptr.String("100%"),
		OwnerID:             ptr.String("123456789012"),
		OwnerAlias:          ptr.String("amazon"),
		Encrypted:           ptr.Bool(true),
		KmsKeyID:            ptr.String("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"),
		DataEncryptionKeyID: ptr.String("12345678-1234-1234-1234-123456789012"),
		StorageTier:         &storageTier,
		RestoreExpiryTime:   &restoreExpiryTime,
		Tags: &[]ec2types.Tag{
			{
				Key:   aws.String("Environment"),
				Value: aws.String("production"),
			},
			{
				Key:   aws.String("Backup"),
				Value: aws.String("daily"),
			},
		},
	}

	properties := ec2Snapshot.Properties()

	a.Equal("snap-1234567890abcdef0", properties.Get("SnapshotID"))
	a.Equal("My snapshot", properties.Get("Description"))
	a.Equal("vol-1234567890abcdef0", properties.Get("VolumeID"))
	a.Equal("100", properties.Get("VolumeSize"))
	a.Equal("completed", properties.Get("State"))
	a.Equal("100% complete", properties.Get("StateMessage"))
	a.Equal("100%", properties.Get("Progress"))
	a.Equal("123456789012", properties.Get("OwnerID"))
	a.Equal("amazon", properties.Get("OwnerAlias"))
	a.Equal("true", properties.Get("Encrypted"))
	a.Equal("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012", properties.Get("KmsKeyID"))
	a.Equal("12345678-1234-1234-1234-123456789012", properties.Get("DataEncryptionKeyID"))
	a.Equal("standard", properties.Get("StorageTier"))
	a.Equal("production", properties.Get("tag:Environment"))
	a.Equal("daily", properties.Get("tag:Backup"))
}

func Test_EC2Snapshot_Properties_EmptyTags(t *testing.T) {
	a := assert.New(t)

	startTime := time.Now()
	state := ec2types.SnapshotStatePending

	ec2Snapshot := EC2Snapshot{
		SnapshotID:  ptr.String("snap-1234567890abcdef0"),
		Description: ptr.String("Automated backup"),
		VolumeID:    ptr.String("vol-1234567890abcdef0"),
		VolumeSize:  ptr.Int32(50),
		State:       &state,
		StartTime:   &startTime,
		Progress:    ptr.String("50%"),
		OwnerID:     ptr.String("123456789012"),
		Encrypted:   ptr.Bool(false),
		Tags:        &[]ec2types.Tag{},
	}

	properties := ec2Snapshot.Properties()

	a.Equal("snap-1234567890abcdef0", properties.Get("SnapshotID"))
	a.Equal("Automated backup", properties.Get("Description"))
	a.Equal("vol-1234567890abcdef0", properties.Get("VolumeID"))
	a.Equal("50", properties.Get("VolumeSize"))
	a.Equal("pending", properties.Get("State"))
	a.Equal("50%", properties.Get("Progress"))
	a.Equal("123456789012", properties.Get("OwnerID"))
	a.Equal("false", properties.Get("Encrypted"))
}

func Test_EC2Snapshot_Properties_SpecialCharactersInTags(t *testing.T) {
	a := assert.New(t)

	startTime := time.Now()
	state := ec2types.SnapshotStateCompleted
	storageTier := ec2types.StorageTierArchive

	ec2Snapshot := EC2Snapshot{
		SnapshotID:  ptr.String("snap-1234567890abcdef0"),
		Description: ptr.String("Weekly backup"),
		VolumeID:    ptr.String("vol-1234567890abcdef0"),
		VolumeSize:  ptr.Int32(200),
		State:       &state,
		StartTime:   &startTime,
		Progress:    ptr.String("100%"),
		OwnerID:     ptr.String("123456789012"),
		Encrypted:   ptr.Bool(true),
		StorageTier: &storageTier,
		Tags: &[]ec2types.Tag{
			{
				Key:   aws.String("Environment:Stage"),
				Value: aws.String("prod/staging"),
			},
			{
				Key:   aws.String("Backup-Schedule"),
				Value: aws.String("weekly/monthly"),
			},
		},
	}

	properties := ec2Snapshot.Properties()

	a.Equal("snap-1234567890abcdef0", properties.Get("SnapshotID"))
	a.Equal("Weekly backup", properties.Get("Description"))
	a.Equal("vol-1234567890abcdef0", properties.Get("VolumeID"))
	a.Equal("200", properties.Get("VolumeSize"))
	a.Equal("completed", properties.Get("State"))
	a.Equal("100%", properties.Get("Progress"))
	a.Equal("123456789012", properties.Get("OwnerID"))
	a.Equal("true", properties.Get("Encrypted"))
	a.Equal("archive", properties.Get("StorageTier"))
	a.Equal("prod/staging", properties.Get("tag:Environment:Stage"))
	a.Equal("weekly/monthly", properties.Get("tag:Backup-Schedule"))
}
