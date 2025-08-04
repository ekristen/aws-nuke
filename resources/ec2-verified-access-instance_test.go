package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessInstance_Properties(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		instance: &ec2.VerifiedAccessInstance{
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
			Description:              ptr.String("Test verified access instance"),
			CreationTime:             ptr.String(now),
			LastUpdatedTime:          ptr.String(now),
			VerifiedAccessTrustProviders: []*ec2.VerifiedAccessTrustProviderCondensed{
				{
					VerifiedAccessTrustProviderId: ptr.String("vatp-1234567890abcdef0"),
				},
				{
					VerifiedAccessTrustProviderId: ptr.String("vatp-1234567890abcdef1"),
				},
			},
			Tags: []*ec2.Tag{
				{
					Key:   ptr.String("Name"),
					Value: ptr.String("TestInstance"),
				},
				{
					Key:   ptr.String("Environment"),
					Value: ptr.String("test"),
				},
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

	// TrustProviders is stored as a string representation of the slice
	assert.NotEmpty(t, properties.Get("TrustProviders"))
}

func Test_EC2VerifiedAccessInstance_Properties_NoTrustProviders(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		instance: &ec2.VerifiedAccessInstance{
			VerifiedAccessInstanceId:     ptr.String("vai-1234567890abcdef0"),
			Description:                  ptr.String("Test verified access instance"),
			CreationTime:                 ptr.String(now),
			LastUpdatedTime:              ptr.String(now),
			VerifiedAccessTrustProviders: nil,
			Tags:                         []*ec2.Tag{},
		},
	}

	properties := instance.Properties()

	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access instance", properties.Get("Description"))
	assert.Equal(t, "", properties.Get("TrustProviders"))
}

func Test_EC2VerifiedAccessInstance_String(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		instance: &ec2.VerifiedAccessInstance{
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
		},
	}

	assert.Equal(t, "vai-1234567890abcdef0", instance.String())
}
