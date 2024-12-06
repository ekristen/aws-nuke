package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockGuardrailResource = "BedrockGuardrail"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockGuardrailResource,
		Scope:    nuke.Account,
		Resource: &BedrockGuardrail{},
		Lister:   &BedrockGuardrailLister{},
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
			tagResp, err := svc.ListTagsForResource(
				&bedrock.ListTagsForResourceInput{
					ResourceARN: guardrail.Arn,
				})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &BedrockGuardrail{
				svc:     svc,
				ID:      guardrail.Id,
				Version: guardrail.Version,
				Name:    guardrail.Name,
				Status:  guardrail.Status,
				Tags:    tagResp.Tags,
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
	svc     *bedrock.Bedrock
	ID      *string
	Version *string
	Name    *string
	Status  *string
	Tags    []*bedrock.Tag
}

func (r *BedrockGuardrail) String() string {
	return *r.ID
}

func (r *BedrockGuardrail) Remove(_ context.Context) error {
	// When guardrail version is not specified, all versions are deleted
	_, err := r.svc.DeleteGuardrail(&bedrock.DeleteGuardrailInput{
		GuardrailIdentifier: r.ID,
	})

	return err
}

func (r *BedrockGuardrail) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
