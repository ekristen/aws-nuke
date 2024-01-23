package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2VPCResource = "EC2VPC"

func init() {
	resource.Register(&resource.Registration{
		Name:   EC2VPCResource,
		Scope:  nuke.Account,
		Lister: &EC2VPCLister{},
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

type EC2VPC struct {
	svc *ec2.EC2
	vpc *ec2.Vpc
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
			svc: svc,
			vpc: vpc,
		})
	}

	return resources, nil
}

func (e *EC2VPC) Remove(_ context.Context) error {
	params := &ec2.DeleteVpcInput{
		VpcId: e.vpc.VpcId,
	}

	_, err := e.svc.DeleteVpc(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2VPC) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.vpc.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("ID", e.vpc.VpcId)
	properties.Set("IsDefault", e.vpc.IsDefault)
	properties.Set("OwnerID", e.vpc.OwnerId)
	return properties
}

func (e *EC2VPC) String() string {
	return *e.vpc.VpcId
}

func DefaultVpc(svc *ec2.EC2) *ec2.Vpc {
	resp, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("is-default"),
				Values: aws.StringSlice([]string{"true"}),
			},
		},
	})
	if err != nil {
		return nil
	}

	if len(resp.Vpcs) == 0 {
		return nil
	}

	return resp.Vpcs[0]
}

func GetVPC(svc *ec2.EC2, vpcID *string) (*ec2.Vpc, error) {
	resp, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{vpcID},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Vpcs) == 0 {
		return nil, nil
	}

	return resp.Vpcs[0], nil
}
