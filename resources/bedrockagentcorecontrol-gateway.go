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
			resources = append(resources, &BedrockAgentCoreGateway{
				svc:            svc,
				GatewayID:      gateway.GatewayId,
				Name:           gateway.Name,
				Status:         string(gateway.Status),
				Description:    gateway.Description,
				AuthorizerType: string(gateway.AuthorizerType),
				ProtocolType:   string(gateway.ProtocolType),
				CreatedAt:      gateway.CreatedAt,
				UpdatedAt:      gateway.UpdatedAt,
			})
		}
	}

	return resources, nil
}

type BedrockAgentCoreGateway struct {
	svc            *bedrockagentcorecontrol.Client
	GatewayID      *string
	Name           *string
	Status         string
	Description    *string
	AuthorizerType string
	ProtocolType   string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

func (r *BedrockAgentCoreGateway) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteGateway(ctx, &bedrockagentcorecontrol.DeleteGatewayInput{
		GatewayIdentifier: r.GatewayID,
	})

	return err
}

func (r *BedrockAgentCoreGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockAgentCoreGateway) String() string {
	return *r.GatewayID
}
