package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseMigrationServiceInstanceProfileProperties(t *testing.T) {
	resource := &DatabaseMigrationServiceInstanceProfile{
		ARN:  ptr.String("arn:aws:dms:us-east-1:123456789012:instance-profile:profile"),
		Name: ptr.String("profile"),
	}

	properties := resource.Properties()

	assert.Equal(t, "arn:aws:dms:us-east-1:123456789012:instance-profile:profile", properties.Get("ARN"))
	assert.Equal(t, "profile", properties.Get("Name"))
	assert.Equal(t, "arn:aws:dms:us-east-1:123456789012:instance-profile:profile", resource.String())
}
