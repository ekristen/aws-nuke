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
		CodeInterpreterID:  ptr.String("test-interpreter-id"),
		CodeInterpreterArn: ptr.String("arn:aws:bedrock:us-east-1:123456789012:code-interpreter/test"),
		Name:               ptr.String("test-interpreter-name"),
		Status:             "ACTIVE",
		Description:        ptr.String("Test code interpreter"),
		CreatedAt:          &createdAt,
		LastUpdatedAt:      &updatedAt,
	}

	props := resource.Properties()

	a.Equal("test-interpreter-id", props.Get("CodeInterpreterID"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:code-interpreter/test", props.Get("CodeInterpreterArn"))
	a.Equal("test-interpreter-name", props.Get("Name"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("Test code interpreter", props.Get("Description"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("LastUpdatedAt"))
}

func Test_BedrockAgentCoreCodeInterpreter_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreCodeInterpreter{
		CodeInterpreterID: ptr.String("test-interpreter-id"),
	}

	a.Equal("test-interpreter-id", resource.String())
}
