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

const BedrockCustomModelResource = "BedrockCustomModel"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockCustomModelResource,
		Scope:  nuke.Account,
		Lister: &BedrockCustomModelLister{},
	})
}

type BedrockCustomModelLister struct{}

func (l *BedrockCustomModelLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrock.ListCustomModelsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListCustomModels(params)
		if err != nil {
			return nil, err
		}

		for _, modelSummary := range resp.ModelSummaries {
			tagResp, err := svc.ListTagsForResource(
				&bedrock.ListTagsForResourceInput{
					ResourceARN: modelSummary.ModelArn,
				})
			if err != nil {
				return nil, err
			}
			resources = append(resources, &BedrockCustomModel{
				svc:  svc,
				Name: modelSummary.ModelName,
				Tags: tagResp.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockCustomModel struct {
	svc  *bedrock.Bedrock
	Name *string
	Tags []*bedrock.Tag
}

func (r *BedrockCustomModel) Remove(_ context.Context) error {
	_, err := r.svc.DeleteCustomModel(&bedrock.DeleteCustomModelInput{
		ModelIdentifier: r.Name,
	})

	return err
}

func (r *BedrockCustomModel) String() string {
	return *r.Name
}

func (r *BedrockCustomModel) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
