package resources

import (
	"context"
	"time"

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
			resources = append(resources, &BedrockAgentCoreMemory{
				svc:       svc,
				MemoryID:  memory.Id,
				Arn:       memory.Arn,
				Status:    string(memory.Status),
				CreatedAt: memory.CreatedAt,
				UpdatedAt: memory.UpdatedAt,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreMemory struct {
	svc       *bedrockagentcorecontrol.Client
	MemoryID  *string
	Arn       *string
	Status    string
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (r *BedrockAgentCoreMemory) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteMemory(ctx, &bedrockagentcorecontrol.DeleteMemoryInput{
		MemoryId: r.MemoryID,
	})

	return err
}

func (r *BedrockAgentCoreMemory) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreMemory) String() string {
	if r.MemoryID != nil {
		return *r.MemoryID
	}
	return *r.Arn
}
