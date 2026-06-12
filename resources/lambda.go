package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type LambdaEventSourceMappingClient interface {
	ListEventSourceMappings(ctx context.Context, params *lambda.ListEventSourceMappingsInput,
		optFns ...func(*lambda.Options)) (*lambda.ListEventSourceMappingsOutput, error)
	ListTags(ctx context.Context, params *lambda.ListTagsInput,
		optFns ...func(*lambda.Options)) (*lambda.ListTagsOutput, error)
	DeleteEventSourceMapping(ctx context.Context, params *lambda.DeleteEventSourceMappingInput,
		optFns ...func(*lambda.Options)) (*lambda.DeleteEventSourceMappingOutput, error)
}
