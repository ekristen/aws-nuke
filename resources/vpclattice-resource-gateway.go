package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	vpclatticeTypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const VPCLatticeResourceGatewayResource = "VPCLatticeResourceGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     VPCLatticeResourceGatewayResource,
		Scope:    nuke.Account,
		Resource: &VPCLatticeResourceGateway{},
		Lister:   &VPCLatticeResourceGatewayLister{},
		DependsOn: []string{
			"VPCLatticeResourceConfiguration",
		},
	})
}

type VPCLatticeResourceGatewayLister struct{}

func (l *VPCLatticeResourceGatewayLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := vpclattice.NewFromConfig(*opts.Config)
	var resources []resource.Resource

	params := &vpclattice.ListResourceGatewaysInput{}
	paginator := vpclattice.NewListResourceGatewaysPaginator(svc, params)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for i := range resp.Items {
			tagResp, err := svc.ListTagsForResource(ctx, &vpclattice.ListTagsForResourceInput{
				ResourceArn: resp.Items[i].Arn,
			})
			if err != nil {
				return nil, err
			}

			resources = append(resources, &VPCLatticeResourceGateway{
				svc:           svc,
				ID:            resp.Items[i].Id,
				Name:          resp.Items[i].Name,
				ARN:           resp.Items[i].Arn,
				Status:        string(resp.Items[i].Status),
				IPAddressType: string(resp.Items[i].IpAddressType),
				CreatedAt:     resp.Items[i].CreatedAt,
				Tags:          tagResp.Tags,
			})
		}
	}

	return resources, nil
}

type VPCLatticeResourceGateway struct {
	svc           *vpclattice.Client
	ID            *string
	Name          *string
	ARN           *string
	Status        string
	IPAddressType string
	CreatedAt     *time.Time
	Tags          map[string]string
}

func (r *VPCLatticeResourceGateway) Filter() error {
	switch vpclatticeTypes.ResourceGatewayStatus(r.Status) {
	case vpclatticeTypes.ResourceGatewayStatusDeleteInProgress,
		vpclatticeTypes.ResourceGatewayStatusCreateInProgress,
		vpclatticeTypes.ResourceGatewayStatusUpdateInProgress:
		return fmt.Errorf("resource gateway is in %s state", r.Status)
	}
	return nil
}

func (r *VPCLatticeResourceGateway) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteResourceGateway(ctx, &vpclattice.DeleteResourceGatewayInput{
		ResourceGatewayIdentifier: r.ID,
	})
	return err
}

func (r *VPCLatticeResourceGateway) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *VPCLatticeResourceGateway) String() string {
	return *r.Name
}
