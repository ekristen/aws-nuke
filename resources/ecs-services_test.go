package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/ecs" //nolint:staticcheck
)

func Test_ECSService_Properties(t *testing.T) {
	r := &ECSService{
		ServiceARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service"),
		ClusterARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"),
		Tags: []*ecs.Tag{
			{
				Key:   ptr.String("Environment"),
				Value: ptr.String("test"),
			},
			{
				Key:   ptr.String("Project"),
				Value: ptr.String("aws-nuke"),
			},
		},
	}

	properties := r.Properties()

	assert.Equal(t, "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service", properties.Get("ServiceARN"))
	assert.Equal(t, "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster", properties.Get("ClusterARN"))
	assert.Equal(t, "test", properties.Get("tag:Environment"))
	assert.Equal(t, "aws-nuke", properties.Get("tag:Project"))
}

func Test_ECSService_String(t *testing.T) {
	r := &ECSService{
		ServiceARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service"),
		ClusterARN: ptr.String("arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"),
	}

	expected := "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service -> " +
		"arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster"
	assert.Equal(t, expected, r.String())
}
