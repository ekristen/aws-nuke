package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

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
		// A harness owns the runtime it runs on and AWS refuses to delete that runtime
		// directly, so the harness has to go first and take the runtime with it.
		DependsOn: []string{
			BedrockAgentCoreHarnessResource,
		},
	})
}

type BedrockAgentCoreAgentRuntimeLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreAgentRuntimeLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource
	var runtimes []*BedrockAgentCoreAgentRuntime

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
			// Get tags for the agent runtime
			var tags map[string]string
			tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
				ResourceArn: runtime.AgentRuntimeArn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for agent runtime: %s", *runtime.AgentRuntimeArn)
			} else {
				tags = tagsResp.Tags
			}

			runtimes = append(runtimes, &BedrockAgentCoreAgentRuntime{
				svc:                 svc,
				AgentRuntimeID:      runtime.AgentRuntimeId,
				AgentRuntimeName:    runtime.AgentRuntimeName,
				AgentRuntimeVersion: runtime.AgentRuntimeVersion,
				Status:              string(runtime.Status),
				Description:         runtime.Description,
				LastUpdatedAt:       runtime.LastUpdatedAt,
				Tags:                tags,
			})
		}
	}

	if len(runtimes) > 0 {
		managed := listHarnessManagedResources(ctx, svc, opts.Logger)

		for _, runtime := range runtimes {
			if harnessID, ok := managed.AgentRuntimes[ptr.ToString(runtime.AgentRuntimeID)]; ok {
				runtime.harness = ptr.String(harnessID)
			}
		}
	}

	for _, runtime := range runtimes {
		resources = append(resources, runtime)
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
	Tags                map[string]string
	harness             *string
}

func (r *BedrockAgentCoreAgentRuntime) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteAgentRuntime(ctx, &bedrockagentcorecontrol.DeleteAgentRuntimeInput{
		AgentRuntimeId: r.AgentRuntimeID,
	})

	return err
}

// Filter skips a runtime that a harness owns. AWS will not delete it directly; it goes
// away when its harness does.
func (r *BedrockAgentCoreAgentRuntime) Filter() error {
	if r.harness != nil {
		return fmt.Errorf("managed by harness %s", *r.harness)
	}

	return nil
}

func (r *BedrockAgentCoreAgentRuntime) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreAgentRuntime) String() string {
	return *r.AgentRuntimeID
}
