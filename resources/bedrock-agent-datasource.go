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

const BedrockDataSourceResource = "BedrockDataSource"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockDataSourceResource,
		Scope:  nuke.Account,
		Lister: &BedrockDataSourceLister{},
	})
}

type BedrockDataSourceLister struct{}

func (l *BedrockDataSourceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// Require to have workspaces identifiers to query components
	knowledgeBaseListResponse, err := ListKnowledgeBaseForDataSource(svc)
	if err != nil {
		return nil, err
	}

	for _, knowledgeBaseResponse := range knowledgeBaseListResponse {
		params := &bedrockagent.ListDataSourcesInput{
			KnowledgeBaseId: knowledgeBaseResponse.KnowledgeBaseId,
			MaxResults:      aws.Int64(25),
		}

		for {
			resp, err := svc.ListDataSources(params)
			if err != nil {
				return nil, err
			}

			for _, item := range resp.DataSourceSummaries {
				resources = append(resources, &BedrockDataSource{
					svc:             svc,
					ID:              item.DataSourceId,
					Name:            item.Name,
					Status:          item.Status,
					KnowledgeBaseID: knowledgeBaseResponse.KnowledgeBaseId,
				})
			}
			if resp.NextToken == nil {
				break
			}
			params.NextToken = resp.NextToken
		}
	}

	return resources, nil
}

// Utility function to list knowledge bases
func ListKnowledgeBaseForDataSource(svc *bedrockagent.BedrockAgent) ([]*bedrockagent.KnowledgeBaseSummary, error) {
	resources := make([]*bedrockagent.KnowledgeBaseSummary, 0)
	params := &bedrockagent.ListKnowledgeBasesInput{
		MaxResults: aws.Int64(25),
	}
	for {
		resp, err := svc.ListKnowledgeBases(params)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resp.KnowledgeBaseSummaries...)
		if resp.NextToken == nil {
			break
		}
		params.NextToken = resp.NextToken
	}
	return resources, nil
}

type BedrockDataSource struct {
	svc             *bedrockagent.BedrockAgent
	ID              *string
	Name            *string
	Status          *string
	KnowledgeBaseID *string
}

func (r *BedrockDataSource) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockDataSource) Remove(_ context.Context) error {
	_, err := r.svc.DeleteDataSource(&bedrockagent.DeleteDataSourceInput{
		DataSourceId:    r.ID,
		KnowledgeBaseId: r.KnowledgeBaseID,
	})

	return err
}

func (r *BedrockDataSource) String() string {
	return *r.ID
}
