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
		MemoryID:  ptr.String("test-memory-id"),
		Arn:       ptr.String("arn:aws:bedrock:us-east-1:123456789012:memory/test"),
		Status:    "ACTIVE",
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-memory-id", props.Get("MemoryID"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:memory/test", props.Get("Arn"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
}

func Test_BedrockAgentCoreMemory_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreMemory{
		MemoryID: ptr.String("test-memory-id"),
	}

	a.Equal("test-memory-id", resource.String())
}
