package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ConnectSecurityProfile_Properties(t *testing.T) {
	a := assert.New(t)

	resource := ConnectSecurityProfile{
		InstanceID: ptr.String("instance-id"),
		ProfileID:  ptr.String("profile-id"),
		Name:       ptr.String("custom-profile"),
		ARN:        ptr.String("arn:aws:connect:us-east-1:123456789012:instance/instance-id/security-profile/profile-id"),
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("instance-id", props.Get("InstanceID"))
	a.Equal("profile-id", props.Get("ProfileID"))
	a.Equal("custom-profile", props.Get("Name"))
	a.Equal("arn:aws:connect:us-east-1:123456789012:instance/instance-id/security-profile/profile-id", props.Get("ARN"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_ConnectSecurityProfile_String(t *testing.T) {
	a := assert.New(t)

	resource := ConnectSecurityProfile{
		Name: ptr.String("custom-profile"),
	}

	a.Equal("custom-profile", resource.String())
}

func Test_ConnectSecurityProfile_Filter_BuiltIn(t *testing.T) {
	a := assert.New(t)

	for _, name := range []string{"Admin", "Agent", "CallCenterManager", "QualityAnalyst"} {
		resource := ConnectSecurityProfile{
			Name: ptr.String(name),
		}

		err := resource.Filter()
		a.NotNil(err)
		a.Contains(err.Error(), "cannot delete built-in security profile")
	}
}

func Test_ConnectSecurityProfile_Filter_Custom(t *testing.T) {
	a := assert.New(t)

	resource := ConnectSecurityProfile{
		Name: ptr.String("CustomProfile"),
	}

	err := resource.Filter()
	a.Nil(err)
}
