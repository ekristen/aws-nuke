package resources

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EC2VerifiedAccessEndpoint_Properties_MinimalData(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		ID:                    ptr.String("vae-1234567890abcdef0"),
		Description:           ptr.String("Test verified access endpoint"),
		CreationTime:          ptr.String(now),
		LastUpdatedTime:       ptr.String(now),
		VerifiedAccessGroupID: ptr.String("vag-1234567890abcdef0"),
		ApplicationDomain:     ptr.String("example.com"),
		EndpointType:          ptr.String("load-balancer"),
		AttachmentType:        ptr.String(""),
		DomainCertificateArn:  nil,
		Tags:                  []ec2types.Tag{},
	}

	properties := endpoint.Properties()

	assert.Equal(t, "vae-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "", properties.Get("AttachmentType"))
	assert.Equal(t, "", properties.Get("DomainCertificateArn"))
}

func Test_EC2VerifiedAccessEndpoint_Properties_WithTags(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		ID:                    ptr.String("vae-1234567890abcdef0"),
		Description:           ptr.String("Test verified access endpoint with tags"),
		CreationTime:          ptr.String(now),
		LastUpdatedTime:       ptr.String(now),
		VerifiedAccessGroupID: ptr.String("vag-1234567890abcdef0"),
		ApplicationDomain:     ptr.String("example.com"),
		EndpointType:          ptr.String("load-balancer"),
		AttachmentType:        ptr.String("vpc"),
		DomainCertificateArn:  ptr.String("arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"),
		Tags: []ec2types.Tag{
			{
				Key:   ptr.String("Name"),
				Value: ptr.String("TestEndpoint"),
			},
			{
				Key:   ptr.String("Environment"),
				Value: ptr.String("test"),
			},
			{
				Key:   ptr.String("Team"),
				Value: ptr.String("security"),
			},
		},
	}

	properties := endpoint.Properties()

	assert.Equal(t, "vae-1234567890abcdef0", properties.Get("ID"))
	assert.Equal(t, "Test verified access endpoint with tags", properties.Get("Description"))
	assert.Equal(t, "vag-1234567890abcdef0", properties.Get("VerifiedAccessGroupID"))
	assert.Equal(t, "example.com", properties.Get("ApplicationDomain"))
	assert.Equal(t, "load-balancer", properties.Get("EndpointType"))
	assert.Equal(t, "vpc", properties.Get("AttachmentType"))
	assert.Equal(t,
		"arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
		properties.Get("DomainCertificateArn"))
	assert.Equal(t, now, properties.Get("CreationTime"))
	assert.Equal(t, now, properties.Get("LastUpdatedTime"))
	assert.Equal(t, "TestEndpoint", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "security", properties.Get("tag:Team"))
}

func Test_EC2VerifiedAccessEndpoint_String(t *testing.T) {
	endpoint := &EC2VerifiedAccessEndpoint{
		ID: ptr.String("vae-1234567890abcdef0"),
	}

	assert.Equal(t, "vae-1234567890abcdef0", endpoint.String())
}
