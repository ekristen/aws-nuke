package resources

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	shieldtypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_ShieldProtectionGroup_String(t *testing.T) {
	a := assert.New(t)

	shieldProtectionGroup := ShieldProtectionGroup{
		ProtectionGroupID: ptr.String("protection-group-id"),
	}

	a.Equal("protection-group-id", shieldProtectionGroup.String())
}

func Test_ShieldProtectionGroup_Properties(t *testing.T) {
	a := assert.New(t)

	aggregation := shieldtypes.ProtectionGroupAggregationSum
	pattern := shieldtypes.ProtectionGroupPatternAll
	resourceType := shieldtypes.ProtectedResourceTypeCloudfrontDistribution

	shieldProtectionGroup := ShieldProtectionGroup{
		ProtectionGroupID:  ptr.String("protection-group-id"),
		Aggregation:        &aggregation,
		Pattern:            &pattern,
		ResourceType:       &resourceType,
		Members:            &[]string{"arn:aws:cloudfront::123456789012:distribution/EXAMPLE123"},
		ProtectionGroupArn: ptr.String("arn:aws:shield::123456789012:protection-group/protection-group-id"),
		Tags: &[]shieldtypes.Tag{
			{
				Key:   aws.String("Environment"),
				Value: aws.String("production"),
			},
			{
				Key:   aws.String("Team"),
				Value: aws.String("security"),
			},
		},
	}

	properties := shieldProtectionGroup.Properties()

	a.Equal("protection-group-id", properties.Get("ProtectionGroupId"))
	a.Equal("SUM", properties.Get("Aggregation"))
	a.Equal("ALL", properties.Get("Pattern"))
	a.Equal("CLOUDFRONT_DISTRIBUTION", properties.Get("ResourceType"))
	a.Equal("arn:aws:shield::123456789012:protection-group/protection-group-id", properties.Get("ProtectionGroupArn"))
	a.Equal("production", properties.Get("tag:Environment"))
	a.Equal("security", properties.Get("tag:Team"))
}

func Test_ShieldProtectionGroup_Properties_EmptyTags(t *testing.T) {
	a := assert.New(t)

	aggregation := shieldtypes.ProtectionGroupAggregationSum
	pattern := shieldtypes.ProtectionGroupPatternAll

	shieldProtectionGroup := ShieldProtectionGroup{
		ProtectionGroupID:  ptr.String("protection-group-id"),
		Aggregation:        &aggregation,
		Pattern:            &pattern,
		Members:            &[]string{},
		ProtectionGroupArn: ptr.String("arn:aws:shield::123456789012:protection-group/protection-group-id"),
		Tags:               &[]shieldtypes.Tag{},
	}

	properties := shieldProtectionGroup.Properties()

	a.Equal("protection-group-id", properties.Get("ProtectionGroupId"))
	a.Equal("SUM", properties.Get("Aggregation"))
	a.Equal("ALL", properties.Get("Pattern"))
	a.Equal("arn:aws:shield::123456789012:protection-group/protection-group-id", properties.Get("ProtectionGroupArn"))
}

func Test_ShieldProtectionGroup_Properties_SpecialCharactersInTags(t *testing.T) {
	a := assert.New(t)

	aggregation := shieldtypes.ProtectionGroupAggregationSum
	pattern := shieldtypes.ProtectionGroupPatternAll

	shieldProtectionGroup := ShieldProtectionGroup{
		ProtectionGroupID:  ptr.String("protection-group-id"),
		Aggregation:        &aggregation,
		Pattern:            &pattern,
		Members:            &[]string{"arn:aws:cloudfront::123456789012:distribution/EXAMPLE123"},
		ProtectionGroupArn: ptr.String("arn:aws:shield::123456789012:protection-group/protection-group-id"),
		Tags: &[]shieldtypes.Tag{
			{
				Key:   aws.String("Environment:Stage"),
				Value: aws.String("prod/staging"),
			},
			{
				Key:   aws.String("Cost-Center"),
				Value: aws.String("security-team"),
			},
		},
	}

	properties := shieldProtectionGroup.Properties()

	a.Equal("protection-group-id", properties.Get("ProtectionGroupId"))
	a.Equal("prod/staging", properties.Get("tag:Environment:Stage"))
	a.Equal("security-team", properties.Get("tag:Cost-Center"))
}
