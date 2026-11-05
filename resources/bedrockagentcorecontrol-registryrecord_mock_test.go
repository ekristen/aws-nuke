package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreRegistryRecord_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreRegistryRecord{
		RecordID:       ptr.String("test-record-id"),
		Name:           ptr.String("test-record"),
		RegistryID:     ptr.String("Olb0p1W8eCXAG35A"),
		RegistryName:   ptr.String("test-registry"),
		DescriptorType: "MCP",
		RecordVersion:  ptr.String("1"),
		Status:         "APPROVED",
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("test-record-id", props.Get("RecordID"))
	a.Equal("test-record", props.Get("Name"))
	a.Equal("Olb0p1W8eCXAG35A", props.Get("RegistryID"))
	a.Equal("test-registry", props.Get("RegistryName"))
	a.Equal("MCP", props.Get("DescriptorType"))
	a.Equal("1", props.Get("RecordVersion"))
	a.Equal("APPROVED", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_BedrockAgentCoreRegistryRecord_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreRegistryRecord{
		RecordID: ptr.String("test-record-id"),
		Name:     ptr.String("test-record"),
	}

	a.Equal("test-record-id", resource.String())
}
