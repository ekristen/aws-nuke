package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
)

// TimestreamInfluxDBAPI defines the interface for Timestream InfluxDB API operations.
// Defined for dependency injection and test mocking.
type TimestreamInfluxDBAPI interface {
	ListDbInstances(ctx context.Context, params *timestreaminfluxdb.ListDbInstancesInput,
		optFns ...func(*timestreaminfluxdb.Options)) (*timestreaminfluxdb.ListDbInstancesOutput, error)
	DeleteDbInstance(ctx context.Context, params *timestreaminfluxdb.DeleteDbInstanceInput,
		optFns ...func(*timestreaminfluxdb.Options)) (*timestreaminfluxdb.DeleteDbInstanceOutput, error)
	ListTagsForResource(ctx context.Context, params *timestreaminfluxdb.ListTagsForResourceInput,
		optFns ...func(*timestreaminfluxdb.Options)) (*timestreaminfluxdb.ListTagsForResourceOutput, error)
}
