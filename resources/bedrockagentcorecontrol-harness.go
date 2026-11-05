package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	agentcoretypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	libtypes "github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const BedrockAgentCoreHarnessResource = "BedrockAgentCoreHarness"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreHarnessResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreHarness{},
		Lister:   &BedrockAgentCoreHarnessLister{},
	})
}

type BedrockAgentCoreHarnessLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreHarnessLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	// A harness runs on an agent runtime, so it can only exist where those do.
	l.SetSupportedRegions(AgentRuntimeSupportedRegions)

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListHarnessesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListHarnessesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, harness := range resp.Harnesses {
			// Get tags for the harness
			var tags map[string]string
			tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
				ResourceArn: harness.Arn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for harness: %s", *harness.Arn)
			} else {
				tags = tagsResp.Tags
			}

			resources = append(resources, &BedrockAgentCoreHarness{
				svc:            svc,
				HarnessID:      harness.HarnessId,
				HarnessName:    harness.HarnessName,
				HarnessVersion: harness.HarnessVersion,
				Status:         string(harness.Status),
				CreatedAt:      harness.CreatedAt,
				LastUpdatedAt:  harness.UpdatedAt,
				Tags:           tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreHarness struct {
	svc            *bedrockagentcorecontrol.Client
	HarnessID      *string
	HarnessName    *string
	HarnessVersion *string
	Status         string
	CreatedAt      *time.Time
	LastUpdatedAt  *time.Time
	Tags           map[string]string
}

func (r *BedrockAgentCoreHarness) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteHarness(ctx, &bedrockagentcorecontrol.DeleteHarnessInput{
		HarnessId: r.HarnessID,
	})

	return err
}

func (r *BedrockAgentCoreHarness) Filter() error {
	if r.Status == string(agentcoretypes.HarnessStatusDeleting) {
		return fmt.Errorf("already being deleted")
	}

	return nil
}

func (r *BedrockAgentCoreHarness) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreHarness) String() string {
	return *r.HarnessID
}
