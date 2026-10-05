package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
)

// SecurityHubV2API defines the interface for the Security Hub V2 API operations.
// Defined for dependency injection and test mocking.
type SecurityHubV2API interface {
	DescribeSecurityHubV2(ctx context.Context, params *securityhub.DescribeSecurityHubV2Input,
		optFns ...func(*securityhub.Options)) (*securityhub.DescribeSecurityHubV2Output, error)
	DisableSecurityHubV2(ctx context.Context, params *securityhub.DisableSecurityHubV2Input,
		optFns ...func(*securityhub.Options)) (*securityhub.DisableSecurityHubV2Output, error)
}
