package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ram"
)

// Interface for resource injection and test mocks
// From https://stackoverflow.com/questions/72235425/simplifying-aws-sdk-go-v2-testing-mocking
type RamAPI interface {
	DeleteResourceShare(ctx context.Context, params *ram.DeleteResourceShareInput,
		optFns ...func(*ram.Options)) (*ram.DeleteResourceShareOutput, error)
	GetResourceShares(ctx context.Context, params *ram.GetResourceSharesInput,
		optFns ...func(*ram.Options)) (*ram.GetResourceSharesOutput, error)
}

type RamClient struct {
	Client *ram.Client
}

func (c *RamClient) DeleteResourceShare(ctx context.Context, params *ram.DeleteResourceShareInput,
	optFns ...func(*ram.Options)) (*ram.DeleteResourceShareOutput, error) {
	return c.Client.DeleteResourceShare(ctx, params, optFns...)
}

func (c *RamClient) GetResourceShares(ctx context.Context, params *ram.GetResourceSharesInput,
	optFns ...func(*ram.Options)) (*ram.GetResourceSharesOutput, error) {
	return c.Client.GetResourceShares(ctx, params, optFns...)
}
