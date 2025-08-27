package resources

import (
	"fmt"
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessInstance_Properties(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		instance: &ec2types.VerifiedAccessInstance{
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
			Description:              ptr.String("Test verified access instance"),
			CreationTime:             ptr.String(now),
			LastUpdatedTime:          ptr.String(now),
			VerifiedAccessTrustProviders: []ec2types.VerifiedAccessTrustProviderCondensed{
				{
					VerifiedAccessTrustProviderId: ptr.String("vatp-1234567890abcdef0"),
				},
				{
					VerifiedAccessTrustProviderId: ptr.String("vatp-1234567890abcdef1"),
				},
			},
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
		},
		ID:              ptr.String("vai-1234567890abcdef0"),
		Description:     ptr.String("Test verified access instance"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
		TrustProviders:  &[]string{"vatp-1234567890abcdef0", "vatp-1234567890abcdef1"},
	}

	properties := instance.Properties()

	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access instance", properties.Get("Description"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "TestInstance", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	fmt.Printf("%v", properties)
}

func Test_EC2VerifiedAccessInstance_Properties_NoTrustProviders(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		instance: &ec2types.VerifiedAccessInstance{
			VerifiedAccessInstanceId:     ptr.String("vai-1234567890abcdef0"),
			Description:                  ptr.String("Test verified access instance"),
			CreationTime:                 ptr.String(now),
			LastUpdatedTime:              ptr.String(now),
			VerifiedAccessTrustProviders: nil,
			Tags:                         []ec2types.Tag{},
		},
		ID:              ptr.String("vai-1234567890abcdef0"),
		Description:     ptr.String("Test verified access instance"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
		TrustProviders:  &[]string{},
	}

	properties := instance.Properties()

	assert.Equal(t, "vai-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access instance", properties.Get("Description"))
}

func Test_EC2VerifiedAccessInstance_String(t *testing.T) {
	instance := &EC2VerifiedAccessInstance{
		instance: &ec2types.VerifiedAccessInstance{
			VerifiedAccessInstanceId: ptr.String("vai-1234567890abcdef0"),
		},
	}

	assert.Equal(t, "vai-1234567890abcdef0", instance.String())
}
