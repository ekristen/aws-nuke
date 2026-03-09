package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	connecttypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
)

func Test_ConnectInstance_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)

	resource := ConnectInstance{
		ID:            ptr.String("test-instance-id"),
		InstanceAlias: ptr.String("test-alias"),
		ARN:           ptr.String("arn:aws:connect:us-east-1:123456789012:instance/test-instance-id"),
		Status:        "ACTIVE",
		CreatedAt:     &createdAt,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("test-instance-id", props.Get("ID"))
	a.Equal("test-alias", props.Get("InstanceAlias"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/test-instance-id", props.Get("ARN"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectInstance_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectInstance{
		ID:            ptr.String("test-instance-id"),
		InstanceAlias: ptr.String("test-alias"),
	}

	a.Equal("test-alias", resource.String())
}

func Test_ConnectInstance_String_NoAlias(t *testing.T) {
	a := assert.New(t)

	resource := ConnectInstance{
		ID: ptr.String("test-instance-id"),
	}

	a.Equal("test-instance-id", resource.String())
}

func Test_ConnectInstance_Filter_CreationInProgress(t *testing.T) {
	a := assert.New(t)

	resource := ConnectInstance{
		ID:     ptr.String("test-instance-id"),
		Status: string(connecttypes.InstanceStatusCreationInProgress),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "instance is being created")
}

func Test_ConnectInstance_Filter_CreationFailed(t *testing.T) {
	a := assert.New(t)

	resource := ConnectInstance{
		ID:     ptr.String("test-instance-id"),
		Status: string(connecttypes.InstanceStatusCreationFailed),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "instance creation failed")
}

func Test_ConnectInstance_Filter_Active(t *testing.T) {
	a := assert.New(t)

	resource := ConnectInstance{
		ID:     ptr.String("test-instance-id"),
		Status: string(connecttypes.InstanceStatusActive),
	}

	err := resource.Filter()
	a.Nil(err)
}
