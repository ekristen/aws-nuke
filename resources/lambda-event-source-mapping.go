package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const LambdaEventSourceMappingResource = "LambdaEventSourceMapping"

func init() {
	registry.Register(&registry.Registration{
		Name:   LambdaEventSourceMappingResource,
		Scope:  nuke.Account,
		Lister: &LambdaEventSourceMappingLister{},
	})
}

type LambdaEventSourceMappingLister struct{}

func (l *LambdaEventSourceMappingLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := lambda.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &lambda.ListEventSourceMappingsInput{}
	for {
		resp, err := svc.ListEventSourceMappings(params)
		if err != nil {
			return nil, err
		}

		for _, mapping := range resp.EventSourceMappings {
			resources = append(resources, &LambdaEventSourceMapping{
				svc:     svc,
				mapping: mapping,
			})
		}

		if resp.NextMarker == nil {
			break
		}

		params.Marker = resp.NextMarker
	}

	return resources, nil
}

type LambdaEventSourceMapping struct {
	svc     *lambda.Lambda
	mapping *lambda.EventSourceMappingConfiguration
}

func (m *LambdaEventSourceMapping) Remove(_ context.Context) error {
	_, err := m.svc.DeleteEventSourceMapping(&lambda.DeleteEventSourceMappingInput{
		UUID: m.mapping.UUID,
	})

	return err
}

func (m *LambdaEventSourceMapping) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("UUID", m.mapping.UUID)
	properties.Set("EventSourceArn", m.mapping.EventSourceArn)
	properties.Set("FunctionArn", m.mapping.FunctionArn)
	properties.Set("State", m.mapping.State)

	return properties
}
