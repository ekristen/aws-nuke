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

const BedrockAgentCoreRegistryResource = "BedrockAgentCoreRegistry"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreRegistryResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreRegistry{},
		Lister:   &BedrockAgentCoreRegistryLister{},
		// A registry holds records, which have to be gone before it can be deleted, and a
		// harness holds the registry, so the harness has to go first as well.
		DependsOn: []string{
			BedrockAgentCoreRegistryRecordResource,
			BedrockAgentCoreHarnessResource,
		},
	})
}

type BedrockAgentCoreRegistryLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreRegistryLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListRegistriesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListRegistriesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, reg := range resp.Registries {
			// Get tags for the registry
			var tags map[string]string
			tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
				ResourceArn: reg.RegistryArn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for registry: %s", *reg.RegistryArn)
			} else {
				tags = tagsResp.Tags
			}

			resources = append(resources, &BedrockAgentCoreRegistry{
				svc:        svc,
				RegistryID: reg.RegistryId,
				Name:       reg.Name,
				Status:     string(reg.Status),
				CreatedAt:  reg.CreatedAt,
				UpdatedAt:  reg.UpdatedAt,
				Tags:       tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreRegistry struct {
	svc        *bedrockagentcorecontrol.Client
	RegistryID *string
	Name       *string
	Status     string
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
	Tags       map[string]string
}

func (r *BedrockAgentCoreRegistry) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteRegistry(ctx, &bedrockagentcorecontrol.DeleteRegistryInput{
		RegistryId: r.RegistryID,
	})

	return err
}

func (r *BedrockAgentCoreRegistry) Filter() error {
	if r.Status == string(agentcoretypes.RegistryStatusDeleting) {
		return fmt.Errorf("already being deleted")
	}

	return nil
}

func (r *BedrockAgentCoreRegistry) Properties() libtypes.Properties {
	return libtypes.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreRegistry) String() string {
	return *r.RegistryID
}
