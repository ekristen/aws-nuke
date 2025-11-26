package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreWorkloadIdentity_Properties(t *testing.T) {
	a := assert.New(t)

	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now()

	resource := BedrockAgentCoreWorkloadIdentity{
		Name:            ptr.String("test-identity-name"),
		CreatedTime:     &createdTime,
		LastUpdatedTime: &updatedTime,
	}

	props := resource.Properties()

	a.Equal("test-identity-name", props.Get("Name"))
	a.Equal(createdTime.Format(time.RFC3339), props.Get("CreatedTime"))
	a.Equal(updatedTime.Format(time.RFC3339), props.Get("LastUpdatedTime"))
}

func Test_BedrockAgentCoreWorkloadIdentity_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreWorkloadIdentity{
		Name: ptr.String("test-identity-name"),
	}

	a.Equal("test-identity-name", resource.String())
}
