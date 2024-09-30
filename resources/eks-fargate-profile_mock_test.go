package resources

import (
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEKSFargateProperties(t *testing.T) {
	resource := &EKSFargateProfile{
		Cluster: ptr.String("test-id"),
		Name:    ptr.String("test-name"),
	}

	properties := resource.Properties()

	assert.Equal(t, "test-id", properties.Get("Cluster"))
	assert.Equal(t, "test-name", properties.Get("Name"))
	assert.Equal(t, "test-id:test-name", resource.String())
}
