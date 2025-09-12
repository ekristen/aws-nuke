package resources

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2Volume_String(t *testing.T) {
	a := assert.New(t)

	ec2Volume := EC2Volume{
		VolumeID: ptr.String("vol-1234567890abcdef0"),
	}

	a.Equal("vol-1234567890abcdef0", ec2Volume.String())
}

func Test_EC2Volume_Properties(t *testing.T) {
	a := assert.New(t)

	createTime := time.Now()
	volumeType := ec2types.VolumeTypeGp3
	state := ec2types.VolumeStateAvailable

	ec2Volume := EC2Volume{
		VolumeID:           ptr.String("vol-1234567890abcdef0"),
		VolumeType:         &volumeType,
		State:              &state,
		Size:               ptr.Int32(100),
		AvailabilityZone:   ptr.String("us-east-1a"),
		CreateTime:         &createTime,
		Encrypted:          ptr.Bool(true),
		KmsKeyID:           ptr.String("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"),
		Iops:               ptr.Int32(3000),
		Throughput:         ptr.Int32(125),
		MultiAttachEnabled: ptr.Bool(false),
		Tags: &[]ec2types.Tag{
			{
				Key:   aws.String("Environment"),
				Value: aws.String("production"),
			},
			{
				Key:   aws.String("Project"),
				Value: aws.String("web-app"),
			},
		},
	}

	properties := ec2Volume.Properties()

	a.Equal("vol-1234567890abcdef0", properties.Get("VolumeID"))
	a.Equal("gp3", properties.Get("VolumeType"))
	a.Equal("available", properties.Get("State"))
	a.Equal("100", properties.Get("Size"))
	a.Equal("us-east-1a", properties.Get("AvailabilityZone"))
	a.Equal("true", properties.Get("Encrypted"))
	a.Equal("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012", properties.Get("KmsKeyID"))
	a.Equal("3000", properties.Get("Iops"))
	a.Equal("125", properties.Get("Throughput"))
	a.Equal("false", properties.Get("MultiAttachEnabled"))
	a.Equal("production", properties.Get("tag:Environment"))
	a.Equal("web-app", properties.Get("tag:Project"))
}

func Test_EC2Volume_Properties_EmptyTags(t *testing.T) {
	a := assert.New(t)

	createTime := time.Now()
	volumeType := ec2types.VolumeTypeGp2
	state := ec2types.VolumeStateInUse

	ec2Volume := EC2Volume{
		VolumeID:         ptr.String("vol-1234567890abcdef0"),
		VolumeType:       &volumeType,
		State:            &state,
		Size:             ptr.Int32(8),
		AvailabilityZone: ptr.String("us-west-2b"),
		CreateTime:       &createTime,
		Encrypted:        ptr.Bool(false),
		Tags:             &[]ec2types.Tag{},
	}

	properties := ec2Volume.Properties()

	a.Equal("vol-1234567890abcdef0", properties.Get("VolumeID"))
	a.Equal("gp2", properties.Get("VolumeType"))
	a.Equal("in-use", properties.Get("State"))
	a.Equal("8", properties.Get("Size"))
	a.Equal("us-west-2b", properties.Get("AvailabilityZone"))
	a.Equal("false", properties.Get("Encrypted"))
}

func Test_EC2Volume_Properties_SpecialCharactersInTags(t *testing.T) {
	a := assert.New(t)

	createTime := time.Now()
	volumeType := ec2types.VolumeTypeIo1
	state := ec2types.VolumeStateAvailable

	ec2Volume := EC2Volume{
		VolumeID:         ptr.String("vol-1234567890abcdef0"),
		VolumeType:       &volumeType,
		State:            &state,
		Size:             ptr.Int32(500),
		AvailabilityZone: ptr.String("eu-west-1c"),
		CreateTime:       &createTime,
		Encrypted:        ptr.Bool(true),
		Iops:             ptr.Int32(5000),
		Tags: &[]ec2types.Tag{
			{
				Key:   aws.String("Environment:Stage"),
				Value: aws.String("prod/staging"),
			},
			{
				Key:   aws.String("Cost-Center"),
				Value: aws.String("dev-team"),
			},
		},
	}

	properties := ec2Volume.Properties()

	a.Equal("vol-1234567890abcdef0", properties.Get("VolumeID"))
	a.Equal("io1", properties.Get("VolumeType"))
	a.Equal("available", properties.Get("State"))
	a.Equal("500", properties.Get("Size"))
	a.Equal("eu-west-1c", properties.Get("AvailabilityZone"))
	a.Equal("true", properties.Get("Encrypted"))
	a.Equal("5000", properties.Get("Iops"))
	a.Equal("prod/staging", properties.Get("tag:Environment:Stage"))
	a.Equal("dev-team", properties.Get("tag:Cost-Center"))
}
