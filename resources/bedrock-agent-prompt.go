package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrockagent"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockPromptResource = "BedrockPrompt"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockPromptResource,
		Scope:  nuke.Account,
		Lister: &BedrockPromptLister{},
	})
}

type BedrockPromptLister struct{}

func (l *BedrockPromptLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrockagent.ListPromptsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListPrompts(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.PromptSummaries {
			resources = append(resources, &BedrockPrompt{
				svc:     svc,
				ID:      item.Id,
				Name:    item.Name,
				Version: item.Version,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockPrompt struct {
	svc     *bedrockagent.BedrockAgent
	ID      *string
	Name    *string
	Version *string
}

func (r *BedrockPrompt) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockPrompt) Remove(_ context.Context) error {
	// Without PromptVersion param, will delete the prompt and all its versions
	_, err := r.svc.DeletePrompt(&bedrockagent.DeletePromptInput{
		PromptIdentifier: r.ID,
	})

	return err
}

func (r *BedrockPrompt) String() string {
	return *r.ID
}
