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
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockAgentCoreMemoryResource = "BedrockAgentCoreMemory"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreMemoryResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreMemory{},
		Lister:   &BedrockAgentCoreMemoryLister{},
	})
}

type BedrockAgentCoreMemoryLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreMemoryLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource
	var memories []*BedrockAgentCoreMemory

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListMemoriesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListMemoriesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, memory := range resp.Memories {
			// Get tags for the memory
			var tags map[string]string
			if memory.Arn != nil {
				tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
					ResourceArn: memory.Arn,
				})
				if err != nil {
					opts.Logger.Warnf("unable to fetch tags for memory: %s", *memory.Arn)
				} else {
					tags = tagsResp.Tags
				}
			}

			memories = append(memories, &BedrockAgentCoreMemory{
				svc:       svc,
				ID:        memory.Id,
				Status:    string(memory.Status),
				CreatedAt: memory.CreatedAt,
				UpdatedAt: memory.UpdatedAt,
				Tags:      tags,
			})
		}
	}

	// Harnesses only exist where agent runtimes do, but memories are listed in more
	// regions than that, so only go looking for owners where one could be.
	if len(memories) > 0 && harnessesSupportedInRegion(opts.Region.Name) {
		managed := listHarnessManagedResources(ctx, svc, opts.Logger)

		for _, memory := range memories {
			if harnessID, ok := managed.Memories[ptr.ToString(memory.ID)]; ok {
				memory.harness = ptr.String(harnessID)
			}
		}
	}

	for _, memory := range memories {
		resources = append(resources, memory)
	}

	return resources, nil
}

type BedrockAgentCoreMemory struct {
	svc       *bedrockagentcorecontrol.Client
	ID        *string
	Status    string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Tags      map[string]string
	harness   *string
}

func (r *BedrockAgentCoreMemory) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteMemory(ctx, &bedrockagentcorecontrol.DeleteMemoryInput{
		MemoryId: r.ID,
	})

	return err
}

// Filter skips a memory that a harness created and manages. AWS will not delete it
// directly; it goes away when its harness does.
func (r *BedrockAgentCoreMemory) Filter() error {
	if r.harness != nil {
		return fmt.Errorf("managed by harness %s", *r.harness)
	}

	return nil
}

func (r *BedrockAgentCoreMemory) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreMemory) String() string {
	return *r.ID
}
