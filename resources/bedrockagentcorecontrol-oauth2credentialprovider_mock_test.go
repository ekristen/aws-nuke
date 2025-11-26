package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreOauth2CredentialProvider_Properties(t *testing.T) {
	a := assert.New(t)

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now()

	resource := BedrockAgentCoreOauth2CredentialProvider{
		Name:                     ptr.String("test-oauth2-provider"),
		CredentialProviderArn:    ptr.String("arn:aws:bedrock:us-east-1:123456789012:oauth2-credential-provider/test"),
		CredentialProviderVendor: "OKTA",
		CreatedTime:              &createdTime,
		LastUpdatedTime:          &updatedTime,
	}

	props := resource.Properties()

	a.Equal("test-oauth2-provider", props.Get("Name"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:oauth2-credential-provider/test", props.Get("CredentialProviderArn"))
	a.Equal("OKTA", props.Get("CredentialProviderVendor"))
	a.Equal(createdTime.Format(time.RFC3339), props.Get("CreatedTime"))
	a.Equal(updatedTime.Format(time.RFC3339), props.Get("LastUpdatedTime"))
}

func Test_BedrockAgentCoreOauth2CredentialProvider_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreOauth2CredentialProvider{
		Name: ptr.String("test-oauth2-provider"),
	}

	a.Equal("test-oauth2-provider", resource.String())
}
