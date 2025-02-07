package resources

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type mockCloudFrontClient struct {
	mock.Mock
}

func (m *mockCloudFrontClient) ListDistributions(ctx context.Context, params *cloudfront.ListDistributionsInput,
	_ ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudfront.ListDistributionsOutput), args.Error(1)
}

func (m *mockCloudFrontClient) ListTagsForResource(ctx context.Context, params *cloudfront.ListTagsForResourceInput,
	_ ...func(*cloudfront.Options)) (*cloudfront.ListTagsForResourceOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudfront.ListTagsForResourceOutput), args.Error(1)
}

func (m *mockCloudFrontClient) GetDistributionConfig(ctx context.Context, params *cloudfront.GetDistributionConfigInput,
	_ ...func(*cloudfront.Options)) (*cloudfront.GetDistributionConfigOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudfront.GetDistributionConfigOutput), args.Error(1)
}

func (m *mockCloudFrontClient) UpdateDistribution(ctx context.Context, params *cloudfront.UpdateDistributionInput,
	_ ...func(*cloudfront.Options)) (*cloudfront.UpdateDistributionOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudfront.UpdateDistributionOutput), args.Error(1)
}

func (m *mockCloudFrontClient) DeleteDistribution(ctx context.Context, params *cloudfront.DeleteDistributionInput,
	_ ...func(*cloudfront.Options)) (*cloudfront.DeleteDistributionOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudfront.DeleteDistributionOutput), args.Error(1)
}
