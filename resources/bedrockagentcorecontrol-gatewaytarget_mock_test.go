package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreGatewayTarget_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreGatewayTarget{
		GatewayIdentifier: ptr.String("test-gateway-id"),
		TargetID:          ptr.String("test-target-id"),
		Name:              ptr.String("test-target-name"),
		Status:            "ACTIVE",
		Description:       ptr.String("Test gateway target"),
		CreatedAt:         &createdAt,
		UpdatedAt:         &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-gateway-id", props.Get("GatewayIdentifier"))
	a.Equal("test-target-id", props.Get("TargetID"))
	a.Equal("test-target-name", props.Get("Name"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("Test gateway target", props.Get("Description"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
}

func Test_BedrockAgentCoreGatewayTarget_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreGatewayTarget{
		TargetID: ptr.String("test-target-id"),
	}

	a.Equal("test-target-id", resource.String())
}
