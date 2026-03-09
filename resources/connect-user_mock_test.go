package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectUser_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectUser{
		InstanceID: ptr.String("instance-id"),
		UserID:     ptr.String("user-id"),
		Username:   ptr.String("testuser"),
		ARN:        ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/agent/user-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("user-id", props.Get("UserID"))
	a.Equal("testuser", props.Get("Username"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/agent/user-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectUser_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectUser{
		Username: ptr.String("testuser"),
	}

	a.Equal("testuser", resource.String())
}
