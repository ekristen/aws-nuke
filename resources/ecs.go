package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// ECSServiceClient is the interface for ECS operations used by ECSService and ECSExpressGatewayService.
type ECSServiceClient interface {
	ListClusters(ctx context.Context, params *ecs.ListClustersInput,
		optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
	ListServices(ctx context.Context, params *ecs.ListServicesInput,
		optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
	ListTagsForResource(ctx context.Context, params *ecs.ListTagsForResourceInput,
		optFns ...func(*ecs.Options)) (*ecs.ListTagsForResourceOutput, error)
	DeleteService(ctx context.Context, params *ecs.DeleteServiceInput,
		optFns ...func(*ecs.Options)) (*ecs.DeleteServiceOutput, error)
	DeleteExpressGatewayService(ctx context.Context, params *ecs.DeleteExpressGatewayServiceInput,
		optFns ...func(*ecs.Options)) (*ecs.DeleteExpressGatewayServiceOutput, error)
}
