package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreCodeInterpreter_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreCodeInterpreter{
		ID:            ptr.String("test-interpreter-id"),
		Name:          ptr.String("test-interpreter-name"),
		Status:        "ACTIVE",
		CreatedAt:     &createdAt,
		LastUpdatedAt: &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-interpreter-id", props.Get("ID"))
	a.Equal("test-interpreter-name", props.Get("Name"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("LastUpdatedAt"))
}

func Test_BedrockAgentCoreCodeInterpreter_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreCodeInterpreter{
		ID: ptr.String("test-interpreter-id"),
	}

	a.Equal("test-interpreter-id", resource.String())
}
