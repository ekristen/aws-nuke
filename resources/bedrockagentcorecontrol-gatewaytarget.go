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

const BedrockAgentCoreGatewayTargetResource = "BedrockAgentCoreGatewayTarget"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreGatewayTargetResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreGatewayTarget{},
		Lister:   &BedrockAgentCoreGatewayTargetLister{},
	})
}

type BedrockAgentCoreGatewayTargetLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreGatewayTargetLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	// First, list all gateways
	gatewayParams := &bedrockagentcorecontrol.ListGatewaysInput{
		MaxResults: aws.Int32(100),
	}

	gatewayPaginator := bedrockagentcorecontrol.NewListGatewaysPaginator(svc, gatewayParams)

	for gatewayPaginator.HasMorePages() {
		gatewayResp, err := gatewayPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// For each gateway, list its targets
		for _, gateway := range gatewayResp.Items {
			targetParams := &bedrockagentcorecontrol.ListGatewayTargetsInput{
				GatewayIdentifier: gateway.GatewayId,
				MaxResults:        aws.Int32(100),
			}

			targetPaginator := bedrockagentcorecontrol.NewListGatewayTargetsPaginator(svc, targetParams)

			for targetPaginator.HasMorePages() {
				targetResp, err := targetPaginator.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, target := range targetResp.Items {
					resources = append(resources, &BedrockAgentCoreGatewayTarget{
						svc:               svc,
						GatewayIdentifier: gateway.GatewayId,
						TargetID:          target.TargetId,
						Name:              target.Name,
						Status:            string(target.Status),
						Description:       target.Description,
						CreatedAt:         target.CreatedAt,
						UpdatedAt:         target.UpdatedAt,
					})
				}
			}
		}
	}

	return resources, nil
}

type BedrockAgentCoreGatewayTarget struct {
	svc               *bedrockagentcorecontrol.Client
	GatewayIdentifier *string
	TargetID          *string
	Name              *string
	Status            string
	Description       *string
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

func (r *BedrockAgentCoreGatewayTarget) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteGatewayTarget(ctx, &bedrockagentcorecontrol.DeleteGatewayTargetInput{
		GatewayIdentifier: r.GatewayIdentifier,
		TargetId:          r.TargetID,
	})

	return err
}

func (r *BedrockAgentCoreGatewayTarget) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreGatewayTarget) String() string {
	return *r.TargetID
}
