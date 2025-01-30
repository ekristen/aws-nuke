package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

var testEC2NetworkInterface = &ec2.NetworkInterface{
	Attachment: &ec2.NetworkInterfaceAttachment{
		AttachmentId: ptr.String("eni-attach-1234567890abcdef0"),
	},
	NetworkInterfaceId: ptr.String("eni-1234567890abcdef0"),
	VpcId:              ptr.String("vpc-12345678"),
	AvailabilityZone:   ptr.String("us-west-2a"),
	OwnerId:            ptr.String("123456789012"),
	PrivateIpAddress:   ptr.String("10.10.10.10"),
	SubnetId:           ptr.String("subnet-12345678"),
	Status:             ptr.String("in-use"),
	TagSet: []*ec2.Tag{
		{
			Key:   ptr.String("Name"),
			Value: ptr.String("test"),
		},
	},
}

func Test_EC2NetworkInterface_Properties(t *testing.T) {
	r := EC2NetworkInterface{
		svc:              nil,
		ID:               testEC2NetworkInterface.NetworkInterfaceId,
		VPC:              testEC2NetworkInterface.VpcId,
		AvailabilityZone: testEC2NetworkInterface.AvailabilityZone,
		PrivateIPAddress: testEC2NetworkInterface.PrivateIpAddress,
		SubnetID:         testEC2NetworkInterface.SubnetId,
		Status:           testEC2NetworkInterface.Status,
		OwnerID:          testEC2NetworkInterface.OwnerId,
		Tags:             testEC2NetworkInterface.TagSet,
		AttachmentID:     testEC2NetworkInterface.Attachment.AttachmentId,
	}

	props := r.Properties()

	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.NetworkInterfaceId), props.Get("ID"))
	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.VpcId), props.Get("VPC"))
	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.AvailabilityZone), props.Get("AvailabilityZone"))
	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.PrivateIpAddress), props.Get("PrivateIPAddress"))
	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.SubnetId), props.Get("SubnetID"))
	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.Status), props.Get("Status"))
	assert.Equal(t, ptr.ToString(testEC2NetworkInterface.OwnerId), props.Get("OwnerID"))
	assert.Equal(t, "test", props.Get("tag:Name"))
}
