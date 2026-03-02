package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_MGNLaunchConfigurationTemplate_Properties_MinimalData(t *testing.T) {
	template := &MGNLaunchConfigurationTemplate{
		LaunchConfigurationTemplateID: ptr.String("lct-1234567890abcdef0"),
		Arn:                           ptr.String("arn:aws:mgn:us-east-1:123456789012:launch-configuration-template/lct-1234567890abcdef0"),
		Tags:                          map[string]string{},
	}

	properties := template.Properties()

	assert.Equal(t, "lct-1234567890abcdef0", properties.Get("LaunchConfigurationTemplateID"))
	assert.Equal(t, "arn:aws:mgn:us-east-1:123456789012:launch-configuration-template/lct-1234567890abcdef0", properties.Get("Arn"))
	assert.Equal(t, "", properties.Get("Ec2LaunchTemplateID"))
	assert.Equal(t, "", properties.Get("LaunchDisposition"))
	assert.Equal(t, "", properties.Get("TargetInstanceTypeRightSizingMethod"))
}

func Test_MGNLaunchConfigurationTemplate_Properties_WithSettings(t *testing.T) {
	template := &MGNLaunchConfigurationTemplate{
		LaunchConfigurationTemplateID:       ptr.String("lct-1234567890abcdef0"),
		Arn:                                 ptr.String("arn:aws:mgn:us-east-1:123456789012:lct/lct-1234567890abcdef0"),
		Ec2LaunchTemplateID:                 ptr.String("lt-1234567890abcdef0"),
		LaunchDisposition:                   "STOPPED",
		TargetInstanceTypeRightSizingMethod: "BASIC",
		CopyPrivateIP:                       ptr.Bool(true),
		CopyTags:                            ptr.Bool(true),
		EnableMapAutoTagging:                ptr.Bool(false),
		Tags: map[string]string{
			"Name":        "TestTemplate",
			"Environment": "test",
			"Purpose":     "migration",
		},
	}

	properties := template.Properties()

	assert.Equal(t, "lct-1234567890abcdef0", properties.Get("LaunchConfigurationTemplateID"))
	assert.Equal(t, "lt-1234567890abcdef0", properties.Get("Ec2LaunchTemplateID"))
	assert.Equal(t, "STOPPED", properties.Get("LaunchDisposition"))
	assert.Equal(t, "BASIC", properties.Get("TargetInstanceTypeRightSizingMethod"))
	assert.Equal(t, "true", properties.Get("CopyPrivateIP"))
	assert.Equal(t, "true", properties.Get("CopyTags"))
	assert.Equal(t, "false", properties.Get("EnableMapAutoTagging"))
	assert.Equal(t, "TestTemplate", properties.Get("tag:Name"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "migration", properties.Get("tag:Purpose"))
}

func Test_MGNLaunchConfigurationTemplate_String(t *testing.T) {
	template := &MGNLaunchConfigurationTemplate{
		LaunchConfigurationTemplateID: ptr.String("lct-1234567890abcdef0"),
	}

	assert.Equal(t, "lct-1234567890abcdef0", template.String())
}
