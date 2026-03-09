package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectContactFlowModule_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlowModule{
		InstanceID: ptr.String("instance-id"),
		ModuleID:   ptr.String("module-id"),
		Name:       ptr.String("test-module"),
		State:      "ACTIVE",
		ARN:        ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/flow-module/module-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("module-id", props.Get("ModuleID"))
	a.Equal("test-module", props.Get("Name"))
	a.Equal("ACTIVE", props.Get("State"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/flow-module/module-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectContactFlowModule_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectContactFlowModule{
		Name: ptr.String("test-module"),
	}

	a.Equal("test-module", resource.String())
}
