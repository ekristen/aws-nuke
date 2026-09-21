package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreHarness_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreHarness{
		HarnessID:      ptr.String("test-harness-id"),
		HarnessName:    ptr.String("test-harness-name"),
		HarnessVersion: ptr.String("1"),
		Status:         "READY",
		CreatedAt:      &createdAt,
		LastUpdatedAt:  &updatedAt,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("test-harness-id", props.Get("HarnessID"))
	a.Equal("test-harness-name", props.Get("HarnessName"))
	a.Equal("1", props.Get("HarnessVersion"))
	a.Equal("READY", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("LastUpdatedAt"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_BedrockAgentCoreHarness_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreHarness{
		HarnessID:   ptr.String("test-harness-id"),
		HarnessName: ptr.String("test-harness-name"),
	}

	a.Equal("test-harness-id", resource.String())
}

func Test_BedrockAgentCoreHarness_Filter(t *testing.T) {
	a := assert.New(t)

	ready := BedrockAgentCoreHarness{
		HarnessID: ptr.String("test-harness-id"),
		Status:    "READY",
	}
	a.Nil(ready.Filter())

	deleting := BedrockAgentCoreHarness{
		HarnessID: ptr.String("test-harness-id"),
		Status:    "DELETING",
	}
	a.EqualError(deleting.Filter(), "already being deleted")
}
