package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreBrowser_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreBrowser{
		BrowserID:     ptr.String("test-browser-id"),
		BrowserArn:    ptr.String("arn:aws:bedrock:us-east-1:123456789012:browser/test"),
		Status:        "ACTIVE",
		Description:   ptr.String("Test browser"),
		CreatedAt:     &createdAt,
		LastUpdatedAt: &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-browser-id", props.Get("BrowserID"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:browser/test", props.Get("BrowserArn"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("Test browser", props.Get("Description"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("LastUpdatedAt"))
}

func Test_BedrockAgentCoreBrowser_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreBrowser{
		BrowserID: ptr.String("test-browser-id"),
	}

	a.Equal("test-browser-id", resource.String())
}
