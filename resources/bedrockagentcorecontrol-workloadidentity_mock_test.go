package resources

import (
	"testing"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
)

func Test_BedrockWorkloadIdentity_Properties(t *testing.T) {
	a := assert.New(t)

	resource := BedrockWorkloadIdentity{
		Name:                ptr.String("test-identity-name"),
		WorkloadIdentityArn: ptr.String("arn:aws:bedrock:us-east-1:123456789012:workload-identity/test"),
	}

	props := resource.Properties()

	a.Equal("test-identity-name", props.Get("Name"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:workload-identity/test", props.Get("WorkloadIdentityArn"))
}

func Test_BedrockWorkloadIdentity_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockWorkloadIdentity{
		WorkloadIdentityArn: ptr.String("arn:aws:bedrock:us-east-1:123456789012:workload-identity/test"),
	}

	a.Equal("arn:aws:bedrock:us-east-1:123456789012:workload-identity/test", resource.String())
}
