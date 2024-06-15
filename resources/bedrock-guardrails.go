package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const BedrockGuardrailResource = "BedrockGuardrail"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockGuardrailResource,
		Scope:  nuke.Account,
		Lister: &BedrockGuardrailLister{},
	})
}

type BedrockGuardrailLister struct{}

func (l *BedrockGuardrailLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrock.ListGuardrailsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListGuardrails(params)
		if err != nil {
			return nil, err
		}

		for _, guardrail := range resp.Guardrails {
			resources = append(resources, &BedrockGuardrail{
				svc:                 svc,
				guardrailIdentifier: guardrail.GuardrailIdentifier,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockGuardrail struct {
	svc                 *bedrock.Bedrock
	guardrailIdentifier *string
}

func (f *BedrockGuardrail) Remove(_ context.Context) error {
	// When guardrail version is not specified, all versions are deleted
	_, err := f.svc.DeleteGuardrail(&bedrock.DeleteGuardrailInput{
		GuardrailIdentifier: f.guardrailIdentifier,
	})

	return err
}

func (f *BedrockGuardrail) String() string {
	return *f.guardrailIdentifier
}

func (f *BedrockGuardrail) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("GuardrailIdentifier", f.guardrailIdentifier)
	return properties
}
