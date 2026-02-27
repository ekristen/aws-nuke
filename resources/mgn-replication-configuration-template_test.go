package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_MGNReplicationConfigurationTemplate_Properties_MinimalData(t *testing.T) {
	template := &MGNReplicationConfigurationTemplate{
		ReplicationConfigurationTemplateID: ptr.String("rct-1234567890abcdef0"),
		ARN:                                ptr.String("arn:aws:mgn:us-east-1:123456789012:rct/rct-1234567890abcdef0"),
		StagingAreaSubnetID:                ptr.String("subnet-1234567890abcdef0"),
		AssociateDefaultSecurityGroup:      ptr.Bool(true),
		BandwidthThrottling:                0,
		CreatePublicIP:                     ptr.Bool(false),
		DataPlaneRouting:                   "PRIVATE_IP",
		DefaultLargeStagingDiskType:        "GP2",
		EBSEncryption:                      "DEFAULT",
		UseDedicatedReplicationServer:      ptr.Bool(false),
		Tags:                               map[string]string{},
	}

	properties := template.Properties()

	assert.Equal(t, "rct-1234567890abcdef0", properties.Get("ReplicationConfigurationTemplateID"))
	assert.Equal(t, "arn:aws:mgn:us-east-1:123456789012:rct/rct-1234567890abcdef0", properties.Get("ARN"))
	assert.Equal(t, "subnet-1234567890abcdef0", properties.Get("StagingAreaSubnetID"))
	assert.Equal(t, "true", properties.Get("AssociateDefaultSecurityGroup"))
	assert.Equal(t, "", properties.Get("BandwidthThrottling"))
	assert.Equal(t, "false", properties.Get("CreatePublicIP"))
	assert.Equal(t, "PRIVATE_IP", properties.Get("DataPlaneRouting"))
	assert.Equal(t, "GP2", properties.Get("DefaultLargeStagingDiskType"))
	assert.Equal(t, "DEFAULT", properties.Get("EBSEncryption"))
	assert.Equal(t, "false", properties.Get("UseDedicatedReplicationServer"))
	assert.Equal(t, "", properties.Get("EbsEncryptionKeyArn"))
	assert.Equal(t, "", properties.Get("ReplicationServerInstanceType"))
}

func Test_MGNReplicationConfigurationTemplate_Properties_WithEncryption(t *testing.T) {
	template := &MGNReplicationConfigurationTemplate{
		ReplicationConfigurationTemplateID: ptr.String("rct-1234567890abcdef0"),
		ARN:                                ptr.String("arn:aws:mgn:us-east-1:123456789012:rct/rct-1234567890abcdef0"),
		StagingAreaSubnetID:                ptr.String("subnet-1234567890abcdef0"),
		AssociateDefaultSecurityGroup:      ptr.Bool(false),
		BandwidthThrottling:                1000,
		CreatePublicIP:                     ptr.Bool(true),
		DataPlaneRouting:                   "PUBLIC_IP",
		DefaultLargeStagingDiskType:        "GP3",
		EBSEncryption:                      "CUSTOM",
		EBSEncryptionKeyARN:                ptr.String("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"),
		ReplicationServerInstanceType:      ptr.String("t3.micro"),
		UseDedicatedReplicationServer:      ptr.Bool(true),
		Tags: map[string]string{
			"Name":        "TestRCT",
			"Environment": "production",
			"Encryption":  "enabled",
		},
	}

	properties := template.Properties()

	assert.Equal(t, "rct-1234567890abcdef0", properties.Get("ReplicationConfigurationTemplateID"))
	assert.Equal(t, "false", properties.Get("AssociateDefaultSecurityGroup"))
	assert.Equal(t, "1000", properties.Get("BandwidthThrottling"))
	assert.Equal(t, "true", properties.Get("CreatePublicIP"))
	assert.Equal(t, "PUBLIC_IP", properties.Get("DataPlaneRouting"))
	assert.Equal(t, "GP3", properties.Get("DefaultLargeStagingDiskType"))
	assert.Equal(t, "CUSTOM", properties.Get("EBSEncryption"))
	assert.Equal(t, "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012", properties.Get("EbsEncryptionKeyArn"))
	assert.Equal(t, "t3.micro", properties.Get("ReplicationServerInstanceType"))
	assert.Equal(t, "true", properties.Get("UseDedicatedReplicationServer"))
	assert.Equal(t, "TestRCT", properties.Get("tag:Name"))
	assert.Equal(t, "production", properties.Get("tag:Environment"))
	assert.Equal(t, "enabled", properties.Get("tag:Encryption"))
}

func Test_MGNReplicationConfigurationTemplate_String(t *testing.T) {
	template := &MGNReplicationConfigurationTemplate{
		ReplicationConfigurationTemplateID: ptr.String("rct-1234567890abcdef0"),
	}

	assert.Equal(t, "rct-1234567890abcdef0", template.String())
}
