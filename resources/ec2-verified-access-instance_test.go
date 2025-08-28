package resources

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessInstance_Properties(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		ID:              ptr.String("vai-1234567890abcdef0"),
		Description:     ptr.String("Test verified access instance"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
		TrustProviders:  &[]string{"vatp-1234567890abcdef0", "vatp-1234567890abcdef1"},
		Tags: []ec2types.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("TestInstance"),
			},
			{
				Key:   ptr.String("Environment"),
				Value: ptr.String("test"),
			},
		},
	}

	properties := instance.Properties()

	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access instance", properties.Get("Description"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "TestInstance", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
}

func Test_EC2VerifiedAccessInstance_Properties_ComprehensiveTags(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		ID:              ptr.String("vai-1234567890abcdef0"),
		Description:     ptr.String("Test verified access instance with comprehensive tags"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
		TrustProviders:  &[]string{"vatp-1234567890abcdef0", "vatp-1234567890abcdef1", "vatp-1234567890abcdef2"},
		Tags: []ec2types.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("ProductionInstance"),
			},
			{
				Key:   ptr.String("Environment"),
				Value: ptr.String("production"),
			},
			{
				Key:   ptr.String("Team"),
				Value: ptr.String("security"),
			},
			{
				Key:   ptr.String("Project"),
				Value: ptr.String("zero-trust"),
			},
			{
				Key:   ptr.String("CostCenter"),
				Value: ptr.String("12345"),
			},
		},
	}

	properties := instance.Properties()

	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access instance with comprehensive tags", properties.Get("Description"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "ProductionInstance", properties.Get("tag:Name"))
	assert.Equal(t, "production", properties.Get("tag:Environment"))
	assert.Equal(t, "security", properties.Get("tag:Team"))
	assert.Equal(t, "zero-trust", properties.Get("tag:Project"))
	assert.Equal(t, "12345", properties.Get("tag:CostCenter"))
}

func Test_EC2VerifiedAccessInstance_Properties_NoTrustProviders(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		ID:              ptr.String("vai-1234567890abcdef0"),
		Description:     ptr.String("Test verified access instance"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
		TrustProviders:  &[]string{},
		Tags:            []ec2types.Tag{},
	}

	properties := instance.Properties()

	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access instance", properties.Get("Description"))
}

func Test_EC2VerifiedAccessInstance_String(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		ID: ptr.String("vai-1234567890abcdef0"),
	}

	assert.Equal(t, "vai-1234567890abcdef0", instance.String())
}
