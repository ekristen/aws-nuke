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

const BedrockGatewayResource = "BedrockGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     BedrockGatewayResource,
		Scope:    nuke.Account,
		Resource: &BedrockGateway{},
		Lister:   &BedrockGatewayLister{},
	})
}

type BedrockGatewayLister struct{}

func (l *BedrockGatewayLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := bedrockagentcorecontrol.NewFromConfig(*opts.Config)
	var resources []resource.Resource

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
			resources = append(resources, &BedrockGateway{
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

type BedrockGateway struct {
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

func (r *BedrockGateway) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteGateway(ctx, &bedrockagentcorecontrol.DeleteGatewayInput{
		GatewayIdentifier: r.GatewayID,
	})

	return err
}

func (r *BedrockGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *BedrockGateway) String() string {
	return *r.GatewayID
}
