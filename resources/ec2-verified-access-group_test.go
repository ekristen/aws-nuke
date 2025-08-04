package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessGroup_Properties(t *testing.T) {
	group := &EC2VerifiedAccessGroup{
		group: &ec2.VerifiedAccessGroup{
			VerifiedAccessGroupId:    ptr.String("vag-1234567890abcdef0"),
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
			Description:              ptr.String("Test verified access group"),
			Owner:                    ptr.String("123456789012"),
			CreationTime:             ptr.String(now),
			LastUpdatedTime:          ptr.String(now),
			Tags: []*ec2.Tag{
				{
					Key:   ptr.String("Name"),
					Value: ptr.String("TestGroup"),
				},
				{
					Key:   ptr.String("Environment"),
					Value: ptr.String("test"),
				},
			},
		},
	}

	properties := group.Properties()

	assert.Equal(t, "vag-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("InstanceID"))
	assert.Equal(t, "Test verified access group", properties.Get("Description"))
	assert.Equal(t, "123456789012", properties.Get("Owner"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "TestGroup", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
}

func Test_EC2VerifiedAccessGroup_String(t *testing.T) {
	group := &EC2VerifiedAccessGroup{
		group: &ec2.VerifiedAccessGroup{
			VerifiedAccessGroupId: ptr.String("vag-1234567890abcdef0"),
		},
	}

	assert.Equal(t, "vag-1234567890abcdef0", group.String())
}
