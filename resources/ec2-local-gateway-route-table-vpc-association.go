package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

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

func (l *EC2LocalGatewayRouteTableVPCAssociationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeLocalGatewayRouteTableVpcAssociations(
		&ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{},
	)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, assoc := range resp.LocalGatewayRouteTableVpcAssociations {
		resources = append(resources, &EC2LocalGatewayRouteTableVPCAssociation{
			svc:   svc,
			assoc: assoc,
		})
	}

	return resources, nil
}

type EC2LocalGatewayRouteTableVPCAssociation struct {
	svc   *ec2.EC2
	assoc *ec2.LocalGatewayRouteTableVpcAssociation
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) Remove(_ context.Context) error {
	_, err := r.svc.DeleteLocalGatewayRouteTableVpcAssociation(
		&ec2.DeleteLocalGatewayRouteTableVpcAssociationInput{
			LocalGatewayRouteTableVpcAssociationId: r.assoc.LocalGatewayRouteTableVpcAssociationId,
		},
	)
	return err
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range r.assoc.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("ID", r.assoc.LocalGatewayRouteTableVpcAssociationId)
	properties.Set("State", r.assoc.State)
	properties.Set("VpcID", r.assoc.VpcId)
	properties.Set("LocalGatewayID", r.assoc.LocalGatewayId)
	properties.Set("LocalGatewayRouteTableID", r.assoc.LocalGatewayRouteTableId)
	properties.Set("OwnerID", r.assoc.OwnerId)
	return properties
}

func (r *EC2LocalGatewayRouteTableVPCAssociation) String() string {
	return ptr.ToString(r.assoc.LocalGatewayRouteTableVpcAssociationId)
}

