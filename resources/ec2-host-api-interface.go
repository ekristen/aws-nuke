package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2HostAPI interface {
	ReleaseHosts(ctx context.Context, params *ec2.ReleaseHostsInput,
		optFns ...func(*ec2.Options)) (*ec2.ReleaseHostsOutput, error)
}
