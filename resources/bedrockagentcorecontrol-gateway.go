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

const BedrockAgentCoreGatewayResource = "BedrockAgentCoreGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockAgentCoreGatewayResource,
		Scope:    nuke.Account,
		Resource: &BedrockAgentCoreGateway{},
		Lister:   &BedrockAgentCoreGatewayLister{},
	})
}

type BedrockAgentCoreGatewayLister struct {
	BedrockAgentCoreControlLister
}

func (l *BedrockAgentCoreGatewayLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	if !l.IsSupportedRegion(opts.Region.Name) {
		return resources, nil
	}

	params := &bedrockagentcorecontrol.ListGatewaysInput{
		MaxResults: aws.Int32(100),
	}

	paginator := bedrockagentcorecontrol.NewListGatewaysPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, gateway := range resp.Items {
			// Get additional gateway details including ARN
			getResp, err := svc.GetGateway(ctx, &bedrockagentcorecontrol.GetGatewayInput{
				GatewayIdentifier: gateway.GatewayId,
			})
			if err != nil {
				return nil, err
			}

			// Get tags for the gateway
			var tags map[string]string
			tagsResp, err := svc.ListTagsForResource(ctx, &bedrockagentcorecontrol.ListTagsForResourceInput{
				ResourceArn: getResp.GatewayArn,
			})
			if err != nil {
				opts.Logger.Warnf("unable to fetch tags for gateway: %s", *getResp.GatewayArn)
			} else {
				tags = tagsResp.Tags
			}

			resources = append(resources, &BedrockAgentCoreGateway{
				svc:            svc,
				ID:             gateway.GatewayId,
				Name:           gateway.Name,
				Status:         string(gateway.Status),
				AuthorizerType: string(gateway.AuthorizerType),
				ProtocolType:   string(gateway.ProtocolType),
				CreatedAt:      gateway.CreatedAt,
				UpdatedAt:      gateway.UpdatedAt,
				Tags:           tags,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreGateway struct {
	svc            *bedrockagentcorecontrol.Client
	ID             *string
	Name           *string
	Status         string
	AuthorizerType string
	ProtocolType   string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	Tags           map[string]string
}

func (r *BedrockAgentCoreGateway) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteGateway(ctx, &bedrockagentcorecontrol.DeleteGatewayInput{
		GatewayIdentifier: r.ID,
	})

	return err
}

func (r *BedrockAgentCoreGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreGateway) String() string {
	return *r.Name
}
