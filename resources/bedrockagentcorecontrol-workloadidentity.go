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

const BedrockAgentCoreWorkloadIdentityResource = "BedrockAgentCoreWorkloadIdentity"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreWorkloadIdentityResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreWorkloadIdentity{},
		Lister:   &BedrockAgentCoreWorkloadIdentityLister{},
	})
}

type BedrockAgentCoreWorkloadIdentityLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreWorkloadIdentityLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

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

			resources = append(resources, &BedrockAgentCoreWorkloadIdentity{
				svc:             svc,
				Name:            identity.Name,
				CreatedTime:     getResp.CreatedTime,
				LastUpdatedTime: getResp.LastUpdatedTime,
				Tags:            tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreWorkloadIdentity struct {
	svc             *bedrockagentcorecontrol.Client
	Name            *string
	CreatedTime     *time.Time
	LastUpdatedTime *time.Time
	Tags            map[string]string
}

func (r *BedrockAgentCoreWorkloadIdentity) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteWorkloadIdentity(ctx, &bedrockagentcorecontrol.DeleteWorkloadIdentityInput{
		Name: r.Name,
	})

	return err
}

func (r *BedrockAgentCoreWorkloadIdentity) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreWorkloadIdentity) String() string {
	return *r.Name
}
