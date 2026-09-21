package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockAgentCoreRegistry_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockAgentCoreRegistry{
		RegistryID: ptr.String("Olb0p1W8eCXAG35A"),
		Name:       ptr.String("test-registry"),
		Status:     "READY",
		CreatedAt:  &createdAt,
		UpdatedAt:  &updatedAt,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("Olb0p1W8eCXAG35A", props.Get("RegistryID"))
	a.Equal("test-registry", props.Get("Name"))
	a.Equal("READY", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_BedrockAgentCoreRegistry_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockAgentCoreRegistry{
		RegistryID: ptr.String("Olb0p1W8eCXAG35A"),
		Name:       ptr.String("test-registry"),
	}

	a.Equal("Olb0p1W8eCXAG35A", resource.String())
}

func Test_BedrockAgentCoreRegistry_Filter(t *testing.T) {
	a := assert.New(t)

	ready := BedrockAgentCoreRegistry{
		RegistryID: ptr.String("Olb0p1W8eCXAG35A"),
		Status:     "READY",
	}
	a.Nil(ready.Filter())

	deleting := BedrockAgentCoreRegistry{
		RegistryID: ptr.String("Olb0p1W8eCXAG35A"),
		Status:     "DELETING",
	}
	a.EqualError(deleting.Filter(), "already being deleted")
}
