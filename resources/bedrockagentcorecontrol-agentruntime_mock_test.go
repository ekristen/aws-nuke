package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreAgentRuntime_Properties(t *testing.T) {
	a := assert.New(t)

	now := time.Now()

	resource := BedrockAgentCoreAgentRuntime{
		AgentRuntimeID:      ptr.String("test-runtime-id"),
		AgentRuntimeName:    ptr.String("test-runtime-name"),
		AgentRuntimeVersion: ptr.String("1.0"),
		Status:              "ACTIVE",
		Description:         ptr.String("Test runtime"),
		LastUpdatedAt:       &now,
	}

	props := resource.Properties()

	a.Equal("test-runtime-id", props.Get("AgentRuntimeID"))
	a.Equal("test-runtime-name", props.Get("AgentRuntimeName"))
	a.Equal("1.0", props.Get("AgentRuntimeVersion"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("Test runtime", props.Get("Description"))
	a.Equal(now.Format(time.RFC3339), props.Get("LastUpdatedAt"))
}

func Test_BedrockAgentCoreAgentRuntime_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreAgentRuntime{
		AgentRuntimeID: ptr.String("test-runtime-id"),
	}

	a.Equal("test-runtime-id", resource.String())
}
