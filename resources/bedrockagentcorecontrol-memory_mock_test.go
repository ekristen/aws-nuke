package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreMemory_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreMemory{
		ID:        ptr.String("test-memory-id"),
		Status:    "ACTIVE",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-memory-id", props.Get("ID"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
}

func Test_BedrockAgentCoreMemory_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreMemory{
		ID: ptr.String("test-memory-id"),
	}

	a.Equal("test-memory-id", resource.String())
}
