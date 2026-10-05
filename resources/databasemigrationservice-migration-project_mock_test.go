package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseMigrationServiceMigrationProjectProperties(t *testing.T) {
	resource := &DatabaseMigrationServiceMigrationProject{
		ARN:  ptr.String("arn:aws:dms:us-east-1:123456789012:migration-project:project"),
		Name: ptr.String("project"),
	}

	properties := resource.Properties()

	assert.Equal(t, "arn:aws:dms:us-east-1:123456789012:migration-project:project", properties.Get("ARN"))
	assert.Equal(t, "project", properties.Get("Name"))
	assert.Equal(t, "arn:aws:dms:us-east-1:123456789012:migration-project:project", resource.String())
}
