package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreGateway_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreGateway{
		GatewayID:      ptr.String("test-gateway-id"),
		Name:           ptr.String("test-gateway-name"),
		Status:         "ACTIVE",
		Description:    ptr.String("Test gateway"),
		AuthorizerType: "CUSTOM_JWT_AUTHORIZER",
		ProtocolType:   "HTTPS",
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-gateway-id", props.Get("GatewayID"))
	a.Equal("test-gateway-name", props.Get("Name"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("Test gateway", props.Get("Description"))
	a.Equal("CUSTOM_JWT_AUTHORIZER", props.Get("AuthorizerType"))
	a.Equal("HTTPS", props.Get("ProtocolType"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
}

func Test_BedrockAgentCoreGateway_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreGateway{
		GatewayID: ptr.String("test-gateway-id"),
	}

	a.Equal("test-gateway-id", resource.String())
}
