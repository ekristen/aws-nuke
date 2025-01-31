package resources

import (
	"context"
	"errors"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VPCResource = "EC2VPC"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VPCResource,
		Scope:    nuke.Account,
		Resource: &EC2VPC{},
		Lister:   &EC2VPCLister{},
		DependsOn: []string{
			EC2SubnetResource,
			EC2RouteTableResource,
			EC2DHCPOptionResource,
			EC2NetworkACLResource,
			EC2NetworkInterfaceResource,
			EC2InternetGatewayAttachmentResource,
			EC2VPCEndpointResource,
			EC2VPCPeeringConnectionResource,
			EC2VPNGatewayResource,
			EC2EgressOnlyInternetGatewayResource,
		},
		AlternativeResource: "AWS::EC2::VPC",
		DeprecatedAliases: []string{
			"EC2Vpc",
		},
	})
}

type EC2VPCLister struct{}

func (l *EC2VPCLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, vpc := range resp.Vpcs {
		resources = append(resources, &EC2VPC{
			svc:       svc,
			vpc:       vpc,
			accountID: opts.AccountID,
		})
	}

	return resources, nil
}

type EC2VPC struct {
	svc       *ec2.EC2
	vpc       *ec2.Vpc
	accountID *string
}

func (r *EC2VPC) Filter() error {
	if ptr.ToString(r.vpc.OwnerId) != ptr.ToString(r.accountID) {
		return errors.New("not owned by account, likely shared")
	}

	return nil
}

func (r *EC2VPC) Remove(_ context.Context) error {
	params := &ec2.DeleteVpcInput{
		VpcId: r.vpc.VpcId,
	}

	_, err := r.svc.DeleteVpc(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2VPC) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range r.vpc.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("ID", r.vpc.VpcId)
	properties.Set("IsDefault", r.vpc.IsDefault)
	properties.Set("OwnerID", r.vpc.OwnerId)
	return properties
}

func (r *EC2VPC) String() string {
	return *r.vpc.VpcId
}
