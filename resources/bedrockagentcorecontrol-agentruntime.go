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

const BedrockAgentCoreAgentRuntimeResource = "BedrockAgentCoreAgentRuntime"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreAgentRuntimeResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreAgentRuntime{},
		Lister:   &BedrockAgentCoreAgentRuntimeLister{},
	})
}

type BedrockAgentCoreAgentRuntimeLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreAgentRuntimeLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	l.SetSupportedRegions(AgentRuntimeSupportedRegions)

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

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
			resources = append(resources, &BedrockAgentCoreAgentRuntime{
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

type BedrockAgentCoreAgentRuntime struct {
	svc                 *bedrockagentcorecontrol.Client
	AgentRuntimeID      *string
	AgentRuntimeName    *string
	AgentRuntimeVersion *string
	Status              string
	Description         *string
	LastUpdatedAt       *time.Time
}

func (r *BedrockAgentCoreAgentRuntime) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAgentRuntime(ctx, &bedrockagentcorecontrol.DeleteAgentRuntimeInput{
		AgentRuntimeId: r.AgentRuntimeID,
	})

	return err
}

func (r *BedrockAgentCoreAgentRuntime) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreAgentRuntime) String() string {
	return *r.AgentRuntimeID
}
