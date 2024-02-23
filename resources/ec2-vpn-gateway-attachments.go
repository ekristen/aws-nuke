package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2VPNGatewayAttachmentResource = "EC2VPNGatewayAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2VPNGatewayAttachmentResource,
		Scope:  nuke.Account,
		Lister: &EC2VPNGatewayAttachmentLister{},
		DeprecatedAliases: []string{
			"EC2VpnGatewayAttachement",
		},
	})
}

type EC2VPNGatewayAttachmentLister struct{}

func (l *EC2VPNGatewayAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, vpc := range resp.Vpcs {
		params := &ec2.DescribeVpnGatewaysInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("attachment.vpc-id"),
					Values: []*string{vpc.VpcId},
				},
			},
		}

		resp, err := svc.DescribeVpnGateways(params)
		if err != nil {
			return nil, err
		}

		for _, vgw := range resp.VpnGateways {
			resources = append(resources, &EC2VPNGatewayAttachment{
				svc:     svc,
				vpcID:   *vpc.VpcId,
				vpnID:   *vgw.VpnGatewayId,
				vpcTags: vpc.Tags,
				vgwTags: vgw.Tags,
			})
		}
	}

	return resources, nil
}

type EC2VPNGatewayAttachment struct {
	svc     *ec2.EC2
	vpcID   string
	vpnID   string
	state   string
	vpcTags []*ec2.Tag
	vgwTags []*ec2.Tag
}

func (v *EC2VPNGatewayAttachment) Filter() error {
	if v.state == "detached" {
		return fmt.Errorf("already detached")
	}
	return nil
}

func (v *EC2VPNGatewayAttachment) Remove(_ context.Context) error {
	params := &ec2.DetachVpnGatewayInput{
		VpcId:        &v.vpcID,
		VpnGatewayId: &v.vpnID,
	}

	_, err := v.svc.DetachVpnGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (v *EC2VPNGatewayAttachment) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range v.vgwTags {
		properties.SetTagWithPrefix("vgw", tagValue.Key, tagValue.Value)
	}
	for _, tagValue := range v.vpcTags {
		properties.SetTagWithPrefix("vpc", tagValue.Key, tagValue.Value)
	}
	return properties
}

func (v *EC2VPNGatewayAttachment) String() string {
	return fmt.Sprintf("%s -> %s", v.vpnID, v.vpcID)
}
