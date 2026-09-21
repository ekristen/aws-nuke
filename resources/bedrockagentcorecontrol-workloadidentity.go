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

const BedrockAgentCoreWorkloadIdentityResource = "BedrockAgentCoreWorkloadIdentity"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreWorkloadIdentityResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreWorkloadIdentity{},
		Lister:   &BedrockAgentCoreWorkloadIdentityLister{},
		// A workload identity created for another AgentCore resource is linked to it and
		// cannot be deleted by the caller, so the owner has to go first and take the
		// identity with it.
		DependsOn: []string{
			BedrockAgentCoreHarnessResource,
			BedrockAgentCoreAgentRuntimeResource,
			BedrockAgentCoreGatewayResource,
			BedrockAgentCoreRegistryResource,
		},
	})
}

type BedrockAgentCoreWorkloadIdentityLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreWorkloadIdentityLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource
	var workloadIdentities []*BedrockAgentCoreWorkloadIdentity

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListWorkloadIdentitiesInput{
		MaxResults: aws.Int32(20),
	}

	paginator := bedrockagentcorecontrol.NewListWorkloadIdentitiesPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, identity := range resp.WorkloadIdentities {
			// Get additional workload identity details
			getResp, err := svc.GetWorkloadIdentity(ctx, &bedrockagentcorecontrol.GetWorkloadIdentityInput{
				Name: identity.Name,
			})
			if err != nil {
				return nil, err
			}

			// Get tags for the workload identity
			var tags map[string]string
			tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
				ResourceArn: identity.WorkloadIdentityArn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for workload identity: %s", *identity.WorkloadIdentityArn)
			} else {
				tags = tagsResp.Tags
			}

			workloadIdentities = append(workloadIdentities, &BedrockAgentCoreWorkloadIdentity{
				svc:             svc,
				Name:            identity.Name,
				CreatedTime:     getResp.CreatedTime,
				LastUpdatedTime: getResp.LastUpdatedTime,
				Tags:            tags,
			})
		}
	}

	if len(workloadIdentities) > 0 {
		owners := workloadIdentityOwners(ctx, svc, opts.Region.Name, opts.Logger)

		for _, identity := range workloadIdentities {
			if owner, ok := owners[ptr.ToString(identity.Name)]; ok {
				identity.owner = ptr.String(owner)
			}
		}
	}

	for _, identity := range workloadIdentities {
		resources = append(resources, identity)
	}

	return resources, nil
}

type BedrockAgentCoreWorkloadIdentity struct {
	svc             *bedrockagentcorecontrol.Client
	Name            *string
	CreatedTime     *time.Time
	LastUpdatedTime *time.Time
	Tags            map[string]string
	owner           *string
}

func (r *BedrockAgentCoreWorkloadIdentity) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteWorkloadIdentity(ctx, &bedrockagentcorecontrol.DeleteWorkloadIdentityInput{
		Name: r.Name,
	})

	return err
}

// Filter skips a workload identity that another AgentCore resource created and owns.
// AWS will not let the caller delete one that is linked to a live service; it goes away
// when its owner does.
func (r *BedrockAgentCoreWorkloadIdentity) Filter() error {
	if r.owner != nil {
		return fmt.Errorf("linked to %s", *r.owner)
	}

	return nil
}

func (r *BedrockAgentCoreWorkloadIdentity) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreWorkloadIdentity) String() string {
	return *r.Name
}
