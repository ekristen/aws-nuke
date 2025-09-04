package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	shieldtypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ShieldProtection_String(t *testing.T) {
	a := assert.New(t)

	shieldProtection := ShieldProtection{
		ID: ptr.String("protection-id"),
	}

	a.Equal("protection-id", shieldProtection.String())
}

func Test_ShieldProtection_Properties(t *testing.T) {
	a := assert.New(t)

	shieldProtection := ShieldProtection{
		ID:            ptr.String("protection-id"),
		Name:          ptr.String("test-protection"),
		ResourceArn:   ptr.String("arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"),
		ProtectionArn: ptr.String("arn:aws:shield::123456789012:protection/protection-id"),
		Tags: &[]shieldtypes.Tag{
			{
				Key:   aws.String("Environment"),
				Value: aws.String("test"),
			},
			{
				Key:   aws.String("Application"),
				Value: aws.String("webapp"),
			},
		},
	}

	properties := shieldProtection.Properties()

	a.Equal("protection-id", properties.Get("ID"))
	a.Equal("test-protection", properties.Get("Name"))
	a.Equal("arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0", properties.Get("ResourceArn"))
	a.Equal("arn:aws:shield::123456789012:protection/protection-id", properties.Get("ProtectionArn"))
	a.Equal("test", properties.Get("tag:Environment"))
	a.Equal("webapp", properties.Get("tag:Application"))
}

func Test_ShieldProtection_Properties_EmptyTags(t *testing.T) {
	a := assert.New(t)

	shieldProtection := ShieldProtection{
		ID:            ptr.String("protection-id"),
		Name:          ptr.String("test-protection"),
		ResourceArn:   ptr.String("arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"),
		ProtectionArn: ptr.String("arn:aws:shield::123456789012:protection/protection-id"),
		Tags:          &[]shieldtypes.Tag{},
	}

	properties := shieldProtection.Properties()

	a.Equal("protection-id", properties.Get("ID"))
	a.Equal("test-protection", properties.Get("Name"))
	a.Equal("arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0", properties.Get("ResourceArn"))
	a.Equal("arn:aws:shield::123456789012:protection/protection-id", properties.Get("ProtectionArn"))
}

func Test_ShieldProtection_Properties_SpecialCharactersInTags(t *testing.T) {
	a := assert.New(t)

	shieldProtection := ShieldProtection{
		ID:            ptr.String("protection-id"),
		Name:          ptr.String("test-protection"),
		ResourceArn:   ptr.String("arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"),
		ProtectionArn: ptr.String("arn:aws:shield::123456789012:protection/protection-id"),
		Tags: &[]shieldtypes.Tag{
			{
				Key:   aws.String("Environment:Test"),
				Value: aws.String("test/value"),
			},
			{
				Key:   aws.String("App-Name"),
				Value: aws.String("web-app"),
			},
		},
	}

	properties := shieldProtection.Properties()

	a.Equal("protection-id", properties.Get("ID"))
	a.Equal("test-protection", properties.Get("Name"))
	a.Equal("test/value", properties.Get("tag:Environment:Test"))
	a.Equal("web-app", properties.Get("tag:App-Name"))
}
