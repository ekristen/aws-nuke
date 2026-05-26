package resources

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type mockECSServiceClient struct {
	mock.Mock
}

func (m *mockECSServiceClient) ListClusters(ctx context.Context, params *ecs.ListClustersInput,
	_ ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.ListClustersOutput), args.Error(1)
}

func (m *mockECSServiceClient) ListServices(ctx context.Context, params *ecs.ListServicesInput,
	_ ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.ListServicesOutput), args.Error(1)
}

func (m *mockECSServiceClient) ListTagsForResource(ctx context.Context, params *ecs.ListTagsForResourceInput,
	_ ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.ListTagsForResourceOutput), args.Error(1)
}

func (m *mockECSServiceClient) DeleteService(ctx context.Context, params *ecs.DeleteServiceInput,
	_ ...func(*ecs.Options)) (*ecs.DeleteServiceOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.DeleteServiceOutput), args.Error(1)
}

func (m *mockECSServiceClient) DeleteExpressGatewayService(ctx context.Context, params *ecs.DeleteExpressGatewayServiceInput,
	_ ...func(*ecs.Options)) (*ecs.DeleteExpressGatewayServiceOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ecs.DeleteExpressGatewayServiceOutput), args.Error(1)
}
