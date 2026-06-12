package resources

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type mockLambdaEventSourceMappingClient struct {
	mock.Mock
}

func (m *mockLambdaEventSourceMappingClient) ListEventSourceMappings(ctx context.Context, params *lambda.ListEventSourceMappingsInput,
	_ ...func(*lambda.Options)) (*lambda.ListEventSourceMappingsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*lambda.ListEventSourceMappingsOutput), args.Error(1)
}

func (m *mockLambdaEventSourceMappingClient) ListTags(ctx context.Context, params *lambda.ListTagsInput,
	_ ...func(*lambda.Options)) (*lambda.ListTagsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*lambda.ListTagsOutput), args.Error(1)
}

func (m *mockLambdaEventSourceMappingClient) DeleteEventSourceMapping(ctx context.Context, params *lambda.DeleteEventSourceMappingInput,
	_ ...func(*lambda.Options)) (*lambda.DeleteEventSourceMappingOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*lambda.DeleteEventSourceMappingOutput), args.Error(1)
}
