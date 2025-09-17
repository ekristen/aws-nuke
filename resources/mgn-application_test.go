package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_MGNApplication_Properties_MinimalData(t *testing.T) {
	application := &MGNApplication{
		ApplicationID:        ptr.String("app-1234567890abcdef0"),
		Arn:                  ptr.String("arn:aws:mgn:us-east-1:123456789012:application/app-1234567890abcdef0"),
		Name:                 ptr.String("TestApplication"),
		Description:          ptr.String("Test migration application"),
		IsArchived:           ptr.Bool(false),
		CreationDateTime:     ptr.String("2024-01-01T12:00:00Z"),
		LastModifiedDateTime: ptr.String("2024-01-01T12:00:00Z"),
		Tags:                 map[string]string{},
	}

	properties := application.Properties()

	assert.Equal(t, "app-1234567890abcdef0", properties.Get("ApplicationID"))
	assert.Equal(t, "arn:aws:mgn:us-east-1:123456789012:application/app-1234567890abcdef0", properties.Get("Arn"))
	assert.Equal(t, "TestApplication", properties.Get("Name"))
	assert.Equal(t, "Test migration application", properties.Get("Description"))
	assert.Equal(t, "false", properties.Get("IsArchived"))
	assert.Equal(t, "2024-01-01T12:00:00Z", properties.Get("CreationDateTime"))
	assert.Equal(t, "2024-01-01T12:00:00Z", properties.Get("LastModifiedDateTime"))
}

func Test_MGNApplication_Properties_WithTags(t *testing.T) {
	application := &MGNApplication{
		ApplicationID:        ptr.String("app-1234567890abcdef0"),
		Arn:                  ptr.String("arn:aws:mgn:us-east-1:123456789012:application/app-1234567890abcdef0"),
		Name:                 ptr.String("WebApplication"),
		Description:          ptr.String("Web application migration"),
		IsArchived:           ptr.Bool(true),
		CreationDateTime:     ptr.String("2024-01-01T12:00:00Z"),
		LastModifiedDateTime: ptr.String("2024-01-01T12:00:00Z"),
		Tags: map[string]string{
			"Name":        "WebApp",
			"Environment": "production",
			"Owner":       "team-web",
			"CostCenter":  "engineering",
		},
	}

	properties := application.Properties()

	assert.Equal(t, "app-1234567890abcdef0", properties.Get("ApplicationID"))
	assert.Equal(t, "WebApplication", properties.Get("Name"))
	assert.Equal(t, "true", properties.Get("IsArchived"))
	assert.Equal(t, "WebApp", properties.Get("tag:Name"))
	assert.Equal(t, "production", properties.Get("tag:Environment"))
	assert.Equal(t, "team-web", properties.Get("tag:Owner"))
	assert.Equal(t, "engineering", properties.Get("tag:CostCenter"))
}

func Test_MGNApplication_String(t *testing.T) {
	application := &MGNApplication{
		ApplicationID: ptr.String("app-1234567890abcdef0"),
	}

	assert.Equal(t, "app-1234567890abcdef0", application.String())
}
