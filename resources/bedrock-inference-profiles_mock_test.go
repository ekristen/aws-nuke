package resources

import (
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
)

func Test_BedrockInferenceProfile_Properties(t *testing.T) {
	a := assert.New(t)

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	resource := BedrockInferenceProfile{
		ID:          ptr.String("test-profile-id"),
		Name:        ptr.String("test-profile-name"),
		ARN:         ptr.String("arn:aws:bedrock:us-east-1:123456789012:inference-profile/test-profile-id"),
		Status:      "ACTIVE",
		Type:        "APPLICATION",
		Description: ptr.String("test description"),
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
		Tags: map[string]string{
			"Environment": "test",
		},
	}

	props := resource.Properties()

	a.Equal("test-profile-id", props.Get("ID"))
	a.Equal("test-profile-name", props.Get("Name"))
	a.Equal("arn:aws:bedrock:us-east-1:123456789012:inference-profile/test-profile-id", props.Get("ARN"))
	a.Equal("ACTIVE", props.Get("Status"))
	a.Equal("APPLICATION", props.Get("Type"))
	a.Equal("test description", props.Get("Description"))
	a.Equal(createdAt.Format(time.RFC3339), props.Get("CreatedAt"))
	a.Equal(updatedAt.Format(time.RFC3339), props.Get("UpdatedAt"))
	a.Equal("test", props.Get("tag:Environment"))
}

func Test_BedrockInferenceProfile_String(t *testing.T) {
	a := assert.New(t)

	resource := BedrockInferenceProfile{
		ID:   ptr.String("test-profile-id"),
		Name: ptr.String("test-profile-name"),
	}

	a.Equal("test-profile-name", resource.String())
}

func Test_BedrockInferenceProfile_Filter_SystemDefined(t *testing.T) {
	a := assert.New(t)

	resource := BedrockInferenceProfile{
		Name: ptr.String("system-profile"),
		Type: string(bedrocktypes.InferenceProfileTypeSystemDefined),
	}

	err := resource.Filter()
	a.NotNil(err)
	a.Contains(err.Error(), "cannot delete system-defined inference profile")
}

func Test_BedrockInferenceProfile_Filter_Application(t *testing.T) {
	a := assert.New(t)

	resource := BedrockInferenceProfile{
		Name: ptr.String("app-profile"),
		Type: string(bedrocktypes.InferenceProfileTypeApplication),
	}

	err := resource.Filter()
	a.Nil(err)
}
