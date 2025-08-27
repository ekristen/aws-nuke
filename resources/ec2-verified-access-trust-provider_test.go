package resources

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessTrustProvider_Properties(t *testing.T) {
	trustProvider := &EC2VerifiedAccessTrustProvider{
		ID:              ptr.String("vatp-1234567890abcdef0"),
		Type:            ptr.String("user"),
		Description:     ptr.String("Test trust provider"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
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

func Test_EC2VerifiedAccessTrustProvider_Properties_SpecialTags(t *testing.T) {
	trustProvider := &EC2VerifiedAccessTrustProvider{
		ID:              ptr.String("vatp-1234567890abcdef0"),
		Type:            ptr.String("device"),
		Description:     ptr.String("Test trust provider with special tags"),
		CreationTime:    ptr.String(now),
		LastUpdatedTime: ptr.String(now),
		Tags: []ec2types.Tag{
			{
				Key:   ptr.String("aws:cloudformation:stack-name"),
				Value: ptr.String("my-stack"),
			},
			{
				Key:   ptr.String("kubernetes.io/cluster/my-cluster"),
				Value: ptr.String("owned"),
			},
			{
				Key:   ptr.String("Project:SubProject"),
				Value: ptr.String("zero-trust:auth"),
			},
		},
	}

	properties := trustProvider.Properties()

	assert.Equal(t, "vatp-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "device", properties.Get("Type"))
	assert.Equal(t, "Test trust provider with special tags", properties.Get("Description"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "my-stack", properties.Get("tag:aws:cloudformation:stack-name"))
	assert.Equal(t, "owned", properties.Get("tag:kubernetes.io/cluster/my-cluster"))
	assert.Equal(t, "zero-trust:auth", properties.Get("tag:Project:SubProject"))
}

func Test_EC2VerifiedAccessTrustProvider_String(t *testing.T) {
	trustProvider := &EC2VerifiedAccessTrustProvider{
		ID: ptr.String("vatp-1234567890abcdef0"),
	}

	assert.Equal(t, "vatp-1234567890abcdef0", trustProvider.String())
}
