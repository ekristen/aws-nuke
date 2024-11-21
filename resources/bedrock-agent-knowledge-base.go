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

const BedrockKnowledgeBaseResource = "BedrockKnowledgeBase"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockKnowledgeBaseResource,
		Scope:  nuke.Account,
		Lister: &BedrockKnowledgeBaseLister{},
		DependsOn: []string{
			BedrockDataSourceResource,
		},
	})
}

type BedrockKnowledgeBaseLister struct{}

func (l *BedrockKnowledgeBaseLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrockagent.ListKnowledgeBasesInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListKnowledgeBases(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.KnowledgeBaseSummaries {
			resources = append(resources, &BedrockKnowledgeBase{
				svc:    svc,
				ID:     item.KnowledgeBaseId,
				Name:   item.Name,
				Status: item.Status,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockKnowledgeBase struct {
	svc    *bedrockagent.BedrockAgent
	ID     *string
	Name   *string
	Status *string
}

func (r *BedrockKnowledgeBase) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockKnowledgeBase) Remove(_ context.Context) error {
	_, err := r.svc.DeleteKnowledgeBase(&bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: r.ID,
	})

	return err
}

func (r *BedrockKnowledgeBase) String() string {
	return *r.ID
}
