package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectRoutingProfile_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectRoutingProfile{
		InstanceID: ptr.String("instance-id"),
		ProfileID:  ptr.String("profile-id"),
		Name:       ptr.String("test-routing-profile"),
		ARN:        ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/routing-profile/profile-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("profile-id", props.Get("ProfileID"))
	a.Equal("test-routing-profile", props.Get("Name"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/routing-profile/profile-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectRoutingProfile_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectRoutingProfile{
		Name: ptr.String("test-routing-profile"),
	}

	a.Equal("test-routing-profile", resource.String())
}

func Test_ConnectRoutingProfile_Filter_Default(t *testing.T) {
	a := assert.New(t)

	resource := ConnectRoutingProfile{
		Name: ptr.String("Basic Routing Profile"),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "cannot delete default routing profile")
}

func Test_ConnectRoutingProfile_Filter_Custom(t *testing.T) {
	a := assert.New(t)

	resource := ConnectRoutingProfile{
		Name: ptr.String("Custom Profile"),
	}

	err := resource.Filter()
	a.Nil(err)
}
