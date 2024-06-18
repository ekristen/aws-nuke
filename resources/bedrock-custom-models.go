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
			resources = append(resources, &BedrockCustomModel{
				svc:       svc,
				modelName: modelSummary.ModelName,
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
	svc       *bedrock.Bedrock
	modelName *string
}

func (f *BedrockCustomModel) Remove(_ context.Context) error {
	_, err := f.svc.DeleteCustomModel(&bedrock.DeleteCustomModelInput{
		ModelIdentifier: f.modelName,
	})

	return err
}

func (f *BedrockCustomModel) String() string {
	return *f.modelName
}

func (f *BedrockCustomModel) Properties() types.Properties {
	properties := types.NewProperties()
	properties.
		Set("ModelName", f.modelName)
	return properties
}
