package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockAgentRuntimeResource = "BedrockAgentRuntime"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentRuntimeResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentRuntime{},
		Lister:   &BedrockAgentRuntimeLister{},
	})
}

type BedrockAgentRuntimeLister struct{}

func (l *BedrockAgentRuntimeLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &bedrockagentcorecontrol.ListAgentRuntimesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, runtime := range resp.AgentRuntimes {
			resources = append(resources, &BedrockAgentRuntime{
				svc:                 svc,
				AgentRuntimeID:      runtime.AgentRuntimeId,
				AgentRuntimeName:    runtime.AgentRuntimeName,
				AgentRuntimeVersion: runtime.AgentRuntimeVersion,
				Status:              string(runtime.Status),
				Description:         runtime.Description,
				LastUpdatedAt:       runtime.LastUpdatedAt,
			})
		}
	}

	return resources, nil
}

type BedrockAgentRuntime struct {
	svc                 *bedrockagentcorecontrol.Client
	AgentRuntimeID      *string
	AgentRuntimeName    *string
	AgentRuntimeVersion *string
	Status              string
	Description         *string
	LastUpdatedAt       *time.Time
}

func (r *BedrockAgentRuntime) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAgentRuntime(ctx, &bedrockagentcorecontrol.DeleteAgentRuntimeInput{
		AgentRuntimeId: r.AgentRuntimeID,
	})

	return err
}

func (r *BedrockAgentRuntime) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentRuntime) String() string {
	return *r.AgentRuntimeID
}
