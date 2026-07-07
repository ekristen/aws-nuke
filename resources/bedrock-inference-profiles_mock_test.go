package resources

import (
	"context"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
)

type mockBedrockInferenceProfileClient struct {
	mock.Mock
}

func (m *mockBedrockInferenceProfileClient) ListInferenceProfiles(ctx context.Context,
	params *bedrock.ListInferenceProfilesInput, _ ...func(*bedrock.Options)) (*bedrock.ListInferenceProfilesOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*bedrock.ListInferenceProfilesOutput), args.Error(1)
}

func (m *mockBedrockInferenceProfileClient) ListTagsForResource(ctx context.Context,
	params *bedrock.ListTagsForResourceInput, _ ...func(*bedrock.Options)) (*bedrock.ListTagsForResourceOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*bedrock.ListTagsForResourceOutput), args.Error(1)
}

func (m *mockBedrockInferenceProfileClient) DeleteInferenceProfile(ctx context.Context,
	params *bedrock.DeleteInferenceProfileInput, _ ...func(*bedrock.Options)) (*bedrock.DeleteInferenceProfileOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*bedrock.DeleteInferenceProfileOutput), args.Error(1)
}

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

// Test_Mock_BedrockInferenceProfile_List_SystemDefined verifies that the lister does not attempt to
// fetch tags for system-defined inference profiles, which do not support tagging. If ListTagsForResource
// were called, the mock would panic because no expectation is registered for it.
func Test_Mock_BedrockInferenceProfile_List_SystemDefined(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockBedrockInferenceProfileClient)

	arn := "arn:aws:bedrock:us-east-2:123456789012:inference-profile/us.anthropic.claude-sonnet-5"
	mockSvc.On("ListInferenceProfiles", mock.Anything, mock.Anything).Return(&bedrock.ListInferenceProfilesOutput{
		InferenceProfileSummaries: []bedrocktypes.InferenceProfileSummary{
			{
				InferenceProfileId:   ptr.String("us.anthropic.claude-sonnet-5"),
				InferenceProfileName: ptr.String("Claude Sonnet 5"),
				InferenceProfileArn:  ptr.String(arn),
				Type:                 bedrocktypes.InferenceProfileTypeSystemDefined,
			},
		},
	}, nil)

	lister := BedrockInferenceProfileLister{mockSvc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	profile := resources[0].(*BedrockInferenceProfile)
	a.Nil(profile.Tags)
	mockSvc.AssertNotCalled(t, "ListTagsForResource", mock.Anything, mock.Anything)
	mockSvc.AssertExpectations(t)
}

// Test_Mock_BedrockInferenceProfile_List_Application verifies that the lister fetches and maps tags
// for application (user-created) inference profiles, which do support tagging.
func Test_Mock_BedrockInferenceProfile_List_Application(t *testing.T) {
	a := assert.New(t)

	mockSvc := new(mockBedrockInferenceProfileClient)

	arn := "arn:aws:bedrock:us-east-2:123456789012:application-inference-profile/abc123"
	mockSvc.On("ListInferenceProfiles", mock.Anything, mock.Anything).Return(&bedrock.ListInferenceProfilesOutput{
		InferenceProfileSummaries: []bedrocktypes.InferenceProfileSummary{
			{
				InferenceProfileId:   ptr.String("abc123"),
				InferenceProfileName: ptr.String("my-app-profile"),
				InferenceProfileArn:  ptr.String(arn),
				Type:                 bedrocktypes.InferenceProfileTypeApplication,
			},
		},
	}, nil)

	mockSvc.On("ListTagsForResource", mock.Anything, &bedrock.ListTagsForResourceInput{
		ResourceARN: ptr.String(arn),
	}).Return(&bedrock.ListTagsForResourceOutput{
		Tags: []bedrocktypes.Tag{
			{Key: ptr.String("Environment"), Value: ptr.String("test")},
		},
	}, nil)

	lister := BedrockInferenceProfileLister{mockSvc: mockSvc}

	resources, err := lister.List(context.TODO(), testListerOpts)
	a.Nil(err)
	a.Len(resources, 1)

	profile := resources[0].(*BedrockInferenceProfile)
	a.Equal("test", profile.Tags["Environment"])
	mockSvc.AssertExpectations(t)
}
