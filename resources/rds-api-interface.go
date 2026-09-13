package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// RDSAPI is the subset of the rds client surface used by the SDK v2 based RDS resources. Defining it as an interface
// lets the listers and resources be exercised with a gomock-generated fake.
type RDSAPI interface {
	DescribeGlobalClusters(ctx context.Context, params *rds.DescribeGlobalClustersInput,
		optFns ...func(*rds.Options)) (*rds.DescribeGlobalClustersOutput, error)
	ModifyGlobalCluster(ctx context.Context, params *rds.ModifyGlobalClusterInput,
		optFns ...func(*rds.Options)) (*rds.ModifyGlobalClusterOutput, error)
	RemoveFromGlobalCluster(ctx context.Context, params *rds.RemoveFromGlobalClusterInput,
		optFns ...func(*rds.Options)) (*rds.RemoveFromGlobalClusterOutput, error)
	DeleteGlobalCluster(ctx context.Context, params *rds.DeleteGlobalClusterInput,
		optFns ...func(*rds.Options)) (*rds.DeleteGlobalClusterOutput, error)
	ListTagsForResource(ctx context.Context, params *rds.ListTagsForResourceInput,
		optFns ...func(*rds.Options)) (*rds.ListTagsForResourceOutput, error)
}
