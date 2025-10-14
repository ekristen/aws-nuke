package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2NATGatewayResource = "EC2NATGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2NATGatewayResource,
		Scope:    nuke.Account,
		Resource: &EC2NATGateway{},
		Lister:   &EC2NATGatewayLister{},
		DeprecatedAliases: []string{
			"EC2NatGateway",
		},
	})
}

type EC2NATGatewayLister struct{}

func (l *EC2NATGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)

	params := &ec2.DescribeNatGatewaysInput{}
	resp, err := svc.DescribeNatGateways(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, natgw := range resp.NatGateways {
		resources = append(resources, &EC2NATGateway{
			svc:   svc,
			natgw: natgw,
		})
	}

	return resources, nil
}

type EC2NATGateway struct {
	svc   *ec2.EC2
	natgw *ec2.NatGateway
}

func (n *EC2NATGateway) Filter() error {
	if *n.natgw.State == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (n *EC2NATGateway) Remove(_ context.Context) error {
	params := &ec2.DeleteNatGatewayInput{
		NatGatewayId: n.natgw.NatGatewayId,
	}

	_, err := n.svc.DeleteNatGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (n *EC2NATGateway) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("CreateTime", n.natgw.CreateTime.Format(time.RFC3339))
	for _, tagValue := range n.natgw.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (n *EC2NATGateway) String() string {
	return ptr.ToString(n.natgw.NatGatewayId)
}
