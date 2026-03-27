package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectQuickConnect_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectQuickConnect{
		InstanceID:       ptr.String("instance-id"),
		QuickConnectID:   ptr.String("qc-id"),
		Name:             ptr.String("test-quick-connect"),
		QuickConnectType: "USER",
		ARN:              ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/transfer-destination/qc-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("qc-id", props.Get("QuickConnectID"))
	a.Equal("test-quick-connect", props.Get("Name"))
	a.Equal("USER", props.Get("QuickConnectType"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/transfer-destination/qc-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectQuickConnect_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectQuickConnect{
		Name: ptr.String("test-quick-connect"),
	}

	a.Equal("test-quick-connect", resource.String())
}
