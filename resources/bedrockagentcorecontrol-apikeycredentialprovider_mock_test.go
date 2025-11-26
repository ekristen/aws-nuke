package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAPIKeyCredentialProvider_Properties(t *testing.T) {
	a := assert.New(t)

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now()

	resource := BedrockAgentCoreAPIKeyCredentialProvider{
		Name:                  ptr.String("test-provider-name"),
		CredentialProviderArn: ptr.String("arn:aws:bedrock:us-east-1:123456789012:credential-provider/test"),
		CreatedTime:           &createdTime,
		LastUpdatedTime:       &updatedTime,
	}

	props := resource.Properties()

	a.Equal("test-provider-name", props.Get("Name"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:credential-provider/test", props.Get("CredentialProviderArn"))
	a.Equal(createdTime.Format(time.RFC3339), props.Get("CreatedTime"))
	a.Equal(updatedTime.Format(time.RFC3339), props.Get("LastUpdatedTime"))
}

func Test_BedrockAPIKeyCredentialProvider_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreAPIKeyCredentialProvider{
		Name: ptr.String("test-provider-name"),
	}

	a.Equal("test-provider-name", resource.String())
}
