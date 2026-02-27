package resources

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessGroup_Properties(t *testing.T) {
	group := &EC2VerifiedAccessGroup{
		ID:                       ptr.String("vag-1234567890abcdef0"),
		Description:              ptr.String("Test verified access group"),
		CreationTime:             ptr.String(now),
		LastUpdatedTime:          ptr.String(now),
		VerifiedAccessInstanceID: ptr.String("vai-1234567890abcdef0"),
		Owner:                    ptr.String("123456789012"),
		Tags: []ec2types.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("TestGroup"),
			},
			{
				Key:   ptr.String("Environment"),
				Value: ptr.String("test"),
			},
		},
	}

	properties := group.Properties()

	assert.Equal(t, "vag-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("VerifiedAccessInstanceID"))
	assert.Equal(t, "Test verified access group", properties.Get("Description"))
	assert.Equal(t, "123456789012", properties.Get("Owner"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "TestGroup", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
}

func Test_EC2VerifiedAccessGroup_Properties_EmptyTags(t *testing.T) {
	group := &EC2VerifiedAccessGroup{
		ID:                       ptr.String("vag-1234567890abcdef0"),
		Description:              ptr.String("Test verified access group without tags"),
		CreationTime:             ptr.String(now),
		LastUpdatedTime:          ptr.String(now),
		VerifiedAccessInstanceID: ptr.String("vai-1234567890abcdef0"),
		Owner:                    ptr.String("123456789012"),
		Tags:                     []ec2types.Tag{},
	}

	properties := group.Properties()

	assert.Equal(t, "vag-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("VerifiedAccessInstanceID"))
	assert.Equal(t, "Test verified access group without tags", properties.Get("Description"))
	assert.Equal(t, "123456789012", properties.Get("Owner"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	// Empty tags should not exist in properties
	assert.Equal(t, "", properties.Get("tag:Name"))
	assert.Equal(t, "", properties.Get("tag:Environment"))
}

func Test_EC2VerifiedAccessGroup_String(t *testing.T) {
	group := &EC2VerifiedAccessGroup{
		ID: ptr.String("vag-1234567890abcdef0"),
	}

	assert.Equal(t, "vag-1234567890abcdef0", group.String())
}
