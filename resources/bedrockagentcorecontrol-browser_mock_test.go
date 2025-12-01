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
		ID:            ptr.String("test-browser-id"),
		Name:          ptr.String("test-browser-name"),
		Status:        "ACTIVE",
		CreatedAt:     &createdAt,
		LastUpdatedAt: &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-browser-id", props.Get("ID"))
	a.Equal("test-browser-name", props.Get("Name"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("LastUpdatedAt"))
}

func Test_BedrockAgentCoreBrowser_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreBrowser{
		ID:   ptr.String("test-browser-id"),
		Name: ptr.String("test-browser-name"),
	}

	a.Equal("test-browser-name", resource.String())
}
