package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestEKSClusterProperties(t *testing.T) {
	resource := &EKSCluster{
		Name: ptr.String("test-cluster"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	properties := resource.Properties()

	assert.Equal(t, "test-cluster", properties.Get("Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "test-cluster", resource.String())
}
