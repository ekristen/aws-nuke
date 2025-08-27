package resources

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessTrustProvider_Properties(t *testing.T) {
	trustProvider := &EC2VerifiedAccessTrustProvider{
		trustProvider: &ec2types.VerifiedAccessTrustProvider{
			VerifiedAccessTrustProviderId: ptr.String("vatp-1234567890abcdef0"),
			TrustProviderType:             "user",
			Description:                   ptr.String("Test trust provider"),
			CreationTime:                  ptr.String(now),
			LastUpdatedTime:               ptr.String(now),
			Tags: []ec2types.Tag{
				{
					Key:   ptr.String("Name"),
					Value: ptr.String("TestTrustProvider"),
				},
				{
					Key:   ptr.String("Environment"),
					Value: ptr.String("test"),
				},
			},
		},
		ID:              ptr.String("vatp-1234567890abcdef0"),
		Type:            ptr.String("user"),
		Description:     ptr.String("Test trust provider"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
	}

	properties := trustProvider.Properties()

	assert.Equal(t, "vatp-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "user", properties.Get("Type"))
	assert.Equal(t, "Test trust provider", properties.Get("Description"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "TestTrustProvider", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
}

func Test_EC2VerifiedAccessTrustProvider_String(t *testing.T) {
	trustProvider := &EC2VerifiedAccessTrustProvider{
		trustProvider: &ec2types.VerifiedAccessTrustProvider{
			VerifiedAccessTrustProviderId: ptr.String("vatp-1234567890abcdef0"),
		},
	}

	assert.Equal(t, "vatp-1234567890abcdef0", trustProvider.String())
}
