package resources

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/emrserverless/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_EMRServerlessApplication_Properties(t *testing.T) {
	now := time.Now()

	app := &EMRServerlessApplication{
		ID:           ptr.String("app-123456"),
		Name:         ptr.String("my-spark-app"),
		Type:         ptr.String("Spark"),
		State:        types.ApplicationStateStarted,
		ARN:          ptr.String("arn:aws:emr-serverless:us-east-1:123456789012:application/app-123456"),
		ReleaseLabel: ptr.String("emr-6.9.0"),
		Architecture: types.ArchitectureX8664,
		CreatedAt:    &now,
		UpdatedAt:    &now,
		Tags: map[string]string{
			"Environment": "dev",
			"Team":        "data-engineering",
		},
	}

	properties := app.Properties()

	assert.Equal(t, "app-123456", properties.Get("ID"))
	assert.Equal(t, "my-spark-app", properties.Get("Name"))
	assert.Equal(t, "Spark", properties.Get("Type"))
	assert.Equal(t, "STARTED", properties.Get("State"))
	assert.Equal(t, "arn:aws:emr-serverless:us-east-1:123456789012:application/app-123456", properties.Get("ARN"))
	assert.Equal(t, "emr-6.9.0", properties.Get("ReleaseLabel"))
	assert.Equal(t, "X86_64", properties.Get("Architecture"))
	assert.Equal(t, now.Format(time.RFC3339), properties.Get("CreatedAt"))
	assert.Equal(t, now.Format(time.RFC3339), properties.Get("UpdatedAt"))
	assert.Equal(t, "dev", properties.Get("tag:Environment"))
	assert.Equal(t, "data-engineering", properties.Get("tag:Team"))
}

func Test_EMRServerlessApplication_String(t *testing.T) {
	app := &EMRServerlessApplication{
		ID:   ptr.String("app-abcd1234"),
		Name: ptr.String("test-application"),
	}

	assert.Equal(t, "app-abcd1234", app.String())
}

func Test_EMRServerlessApplication_Filter_Terminated(t *testing.T) {
	app := &EMRServerlessApplication{
		ID:    ptr.String("app-123"),
		State: types.ApplicationStateTerminated,
	}

	err := app.Filter()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already terminated")
}

func Test_EMRServerlessApplication_Filter_NotTerminated(t *testing.T) {
	app := &EMRServerlessApplication{
		ID:    ptr.String("app-123"),
		State: types.ApplicationStateStarted,
	}

	err := app.Filter()
	assert.NoError(t, err)
}
