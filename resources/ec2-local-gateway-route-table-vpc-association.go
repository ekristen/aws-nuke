package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2LocalGatewayRouteTableVPCAssociationResource = "EC2LocalGatewayRouteTableVPCAssociation"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2LocalGatewayRouteTableVPCAssociationResource,
		Scope:    nuke.Account,
		Resource: &EC2LocalGatewayRouteTableVPCAssociation{},
		Lister:   &EC2LocalGatewayRouteTableVPCAssociationLister{},
	})
}

type EC2LocalGatewayRouteTableVPCAssociationLister struct{}

func (l *EC2LocalGatewayRouteTableVPCAssociationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.NewFromConfig(*opts.Config)

	resp, err := svc.DescribeLocalGatewayRouteTableVpcAssociations(ctx,
		&ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{},
	)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for i := range resp.LocalGatewayRouteTableVpcAssociations {
		assoc := &resp.LocalGatewayRouteTableVpcAssociations[i]
		resources = append(resources, &EC2LocalGatewayRouteTableVPCAssociation{
			svc:                      svc,
			ID:                       assoc.LocalGatewayRouteTableVpcAssociationId,
			State:                    assoc.State,
			VpcID:                    assoc.VpcId,
			LocalGatewayID:           assoc.LocalGatewayId,
			LocalGatewayRouteTableID: assoc.LocalGatewayRouteTableId,
			OwnerID:                  assoc.OwnerId,
			Tags:                     assoc.Tags,
		})
	}

	return resources, nil
}

type EC2LocalGatewayRouteTableVPCAssociation struct {
	svc                      *ec2.Client
	ID                       *string `property:"name=ID"`
	State                    *string
	VpcID                    *string
	LocalGatewayID           *string
	LocalGatewayRouteTableID *string
	OwnerID                  *string
	Tags                     []ec2types.Tag
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) Filter() error {
	if r.OwnerID != nil && strings.HasSuffix(ptr.ToString(r.OwnerID), ".amazonaws.com") {
		return fmt.Errorf("unable to remove service-managed association (owner: %s)", ptr.ToString(r.OwnerID))
	}
	return nil
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteLocalGatewayRouteTableVpcAssociation(ctx,
		&ec2.DeleteLocalGatewayRouteTableVpcAssociationInput{
			LocalGatewayRouteTableVpcAssociationId: r.ID,
		},
	)
	return err
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) String() string {
	return ptr.ToString(r.ID)
}
