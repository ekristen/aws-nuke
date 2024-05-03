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

const BedrockProvisionedModelThroughputResource = "BedrockProvisionedModelThroughput"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockProvisionedModelThroughputResource,
		Scope:  nuke.Account,
		Lister: &BedrockProvisionedModelThroughputLister{},
	})
}

type BedrockProvisionedModelThroughputLister struct{}

func (l *BedrockProvisionedModelThroughputLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrock.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrock.ListProvisionedModelThroughputsInput{
		MaxResults: aws.Int64(30),
	}

	for {
		resp, err := svc.ListProvisionedModelThroughputs(params)
		if err != nil {
			return nil, err
		}

		for _, provisionedModelSummary := range resp.ProvisionedModelSummaries {
			tagResp, err := svc.ListTagsForResource(
				&bedrock.ListTagsForResourceInput{
					ResourceARN: provisionedModelSummary.ProvisionedModelArn,
				})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &BedrockProvisionedModelThroughput{
				svc:    svc,
				Arn:    provisionedModelSummary.ProvisionedModelArn,
				Name:   provisionedModelSummary.ProvisionedModelName,
				Status: provisionedModelSummary.Status,
				Tags:   tagResp.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockProvisionedModelThroughput struct {
	svc    *bedrock.Bedrock
	Arn    *string
	Name   *string
	Status *string
	Tags   []*bedrock.Tag
}

func (r *BedrockProvisionedModelThroughput) Remove(_ context.Context) error {
	_, err := r.svc.DeleteProvisionedModelThroughput(&bedrock.DeleteProvisionedModelThroughputInput{
		ProvisionedModelId: r.Arn,
	})

	return err
}

func (r *BedrockProvisionedModelThroughput) String() string {
	return *r.Name
}

func (r *BedrockProvisionedModelThroughput) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}
