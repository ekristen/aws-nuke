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
		Name:   BedrockAgentAliasResource,
		Scope:  nuke.Account,
		Lister: &BedrockAgentAliasLister{},
	})
}

type BedrockAgentAliasLister struct{}

func (l *BedrockAgentAliasLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	agentIds, err := ListBedrockAgentIds(svc)
	if err != nil {
		return nil, err
	}

	for _, agentId := range agentIds {
		params := &bedrockagent.ListAgentAliasesInput{
			MaxResults: aws.Int64(100),
			AgentId:    aws.String(agentId),
		}
		for {
			output, err := svc.ListAgentAliases(params)
			if err != nil {
				return nil, err
			}

			for _, agentAliasInfo := range output.AgentAliasSummaries {
				resources = append(resources, &BedrockAgentAlias{
					svc:            svc,
					AgentId:        aws.String(agentId),
					AgentAliasName: agentAliasInfo.AgentAliasName,
					AgentAliasId:   agentAliasInfo.AgentAliasId,
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
	AgentId        *string
	AgentAliasId   *string
	AgentAliasName *string
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

func (f *BedrockAgentAlias) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}

func (f *BedrockAgentAlias) Remove(_ context.Context) error {
	_, err := f.svc.DeleteAgentAlias(&bedrockagent.DeleteAgentAliasInput{
		AgentAliasId: f.AgentAliasId,
		AgentId:      f.AgentId,
	})
	return err
}

func (f *BedrockAgentAlias) String() string {
	return *f.AgentAliasName
}
