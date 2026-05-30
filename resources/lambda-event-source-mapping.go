package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LambdaEventSourceMappingResource = "LambdaEventSourceMapping"

func init() {
	registry.Register(&registry.Registration{
		Name:     LambdaEventSourceMappingResource,
		Scope:    nuke.Account,
		Resource: &LambdaEventSourceMapping{},
		Lister:   &LambdaEventSourceMappingLister{},
	})
}

type LambdaEventSourceMappingLister struct {
	mockSvc LambdaEventSourceMappingClient
}

func (l *LambdaEventSourceMappingLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	var svc LambdaEventSourceMappingClient
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = lambda.NewFromConfig(*opts.Config)
	}

	resources := make([]resource.Resource, 0)

	params := &lambda.ListEventSourceMappingsInput{}
	paginator := lambda.NewListEventSourceMappingsPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// Use index-based iteration to avoid copying the 344-byte
		// EventSourceMappingConfiguration struct on each iteration.
		for i := range resp.EventSourceMappings {
			mapping := &resp.EventSourceMappings[i]
			tagsResp, err := svc.ListTags(ctx, &lambda.ListTagsInput{
				Resource: mapping.EventSourceMappingArn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &LambdaEventSourceMapping{
				svc:                   svc,
				UUID:                  mapping.UUID,
				EventSourceMappingArn: mapping.EventSourceMappingArn,
				EventSourceArn:        mapping.EventSourceArn,
				FunctionArn:           mapping.FunctionArn,
				State:                 mapping.State,
				Tags:                  tagsResp.Tags,
			})
		}
	}

	return resources, nil
}

type LambdaEventSourceMapping struct {
	svc                   LambdaEventSourceMappingClient
	UUID                  *string
	EventSourceMappingArn *string
	EventSourceArn        *string
	FunctionArn           *string
	State                 *string
	Tags                  map[string]string
}

func (r *LambdaEventSourceMapping) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteEventSourceMapping(ctx, &lambda.DeleteEventSourceMappingInput{
		UUID: r.UUID,
	})

	return err
}

func (r *LambdaEventSourceMapping) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *LambdaEventSourceMapping) String() string {
	return *r.UUID
}
