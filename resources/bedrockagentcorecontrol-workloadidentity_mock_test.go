package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreWorkloadIdentity_Properties(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreWorkloadIdentity{
		Name: ptr.String("test-identity-name"),
	}

	props := resource.Properties()

	a.Equal("test-identity-name", props.Get("Name"))
}

func Test_BedrockAgentCoreWorkloadIdentity_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreWorkloadIdentity{
		Name: ptr.String("test-identity-name"),
	}

	a.Equal("test-identity-name", resource.String())
}
