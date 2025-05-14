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

const BedrockAgentAliasResource = "BedrockAgentAlias"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentAliasResource,
		Scope:    nuke.Account,
		Lister:   &BedrockAgentAliasLister{},
		Resource: &BedrockAgentAlias{},
		DependsOn: []string{
			BedrockAgentResource,
		},
	})
}

type BedrockAgentAliasLister struct{}

func (l *BedrockAgentAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	agentIDs, err := ListBedrockAgentIds(svc)
	if err != nil {
		return nil, err
	}

	for _, agentID := range agentIDs {
		params := &bedrockagent.ListAgentAliasesInput{
			MaxResults: aws.Int64(100),
			AgentId:    aws.String(agentID),
		}
		for {
			output, err := svc.ListAgentAliases(params)
			if err != nil {
				return nil, err
			}

			for _, agentAliasInfo := range output.AgentAliasSummaries {
				resources = append(resources, &BedrockAgentAlias{
					svc:            svc,
					AgentID:        aws.String(agentID),
					AgentAliasName: agentAliasInfo.AgentAliasName,
					AgentAliasID:   agentAliasInfo.AgentAliasId,
				})
			}

			if output.NextToken == nil {
				break
			}
			params.NextToken = output.NextToken
		}
	}
	return resources, nil
}

type BedrockAgentAlias struct {
	svc            *bedrockagent.BedrockAgent
	AgentID        *string
	AgentAliasID   *string
	AgentAliasName *string
}

func (r *BedrockAgentAlias) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentAlias) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAgentAlias(&bedrockagent.DeleteAgentAliasInput{
		AgentAliasId: r.AgentAliasID,
		AgentId:      r.AgentID,
	})
	return err
}

func (r *BedrockAgentAlias) String() string {
	return *r.AgentAliasName
}

func ListBedrockAgentIds(svc *bedrockagent.BedrockAgent) ([]string, error) {
	agentIds := []string{}
	params := &bedrockagent.ListAgentsInput{
		MaxResults: aws.Int64(100),
	}
	for {
		output, err := svc.ListAgents(params)
		if err != nil {
			return nil, err
		}

		for _, agent := range output.AgentSummaries {
			agentIds = append(agentIds, *agent.AgentId)
		}

		if output.NextToken == nil {
			break
		}
		params.NextToken = output.NextToken
	}

	return agentIds, nil
}
