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

const BedrockAgentResource = "BedrockAgent"

func init() {
	registry.Register(&registry.Registration{
		Name:   BedrockAgentResource,
		Scope:  nuke.Account,
		Lister: &BedrockAgentLister{},
	})
}

type BedrockAgentLister struct{}

func (l *BedrockAgentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := bedrockagent.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &bedrockagent.ListAgentsInput{
		MaxResults: aws.Int64(25),
	}

	for {
		resp, err := svc.ListAgents(params)
		if err != nil {
			return nil, err
		}

		for _, item := range resp.AgentSummaries {
			// Cannot query tags directly here, AgentSummaries do not contain agent ARN...

			resources = append(resources, &BedrockAgent{
				svc:    svc,
				ID:     item.AgentId,
				Name:   item.AgentName,
				Status: item.AgentStatus,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type BedrockAgent struct {
	svc    *bedrockagent.BedrockAgent
	ID     *string
	Name   *string
	Status *string
}

func (r *BedrockAgent) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgent) Remove(_ context.Context) error {
	_, err := r.svc.DeleteAgent(&bedrockagent.DeleteAgentInput{
		AgentId:                r.ID,
		SkipResourceInUseCheck: aws.Bool(true),
	})

	return err
}

func (r *BedrockAgent) String() string {
	return *r.ID
}
