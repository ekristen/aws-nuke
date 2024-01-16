package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2VPNGatewayResource = "EC2VPNGateway"

func init() {
	resource.Register(resource.Registration{
		Name:   EC2VPNGatewayResource,
		Scope:  nuke.Account,
		Lister: &EC2VPNGatewayLister{},
		DependsOn: []string{
			EC2VPNGatewayAttachmentResource,
		},
	})
}

type EC2VPNGatewayLister struct{}

func (l *EC2VPNGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)

	params := &ec2.DescribeVpnGatewaysInput{}
	resp, err := svc.DescribeVpnGateways(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.VpnGateways {
		resources = append(resources, &EC2VPNGateway{
			svc:   svc,
			id:    *out.VpnGatewayId,
			state: *out.State,
		})
	}

	return resources, nil
}

type EC2VPNGateway struct {
	svc   *ec2.EC2
	id    string
	state string
}

func (v *EC2VPNGateway) Filter() error {
	if v.state == "deleted" {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (v *EC2VPNGateway) Remove(_ context.Context) error {
	params := &ec2.DeleteVpnGatewayInput{
		VpnGatewayId: &v.id,
	}

	_, err := v.svc.DeleteVpnGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (v *EC2VPNGateway) String() string {
	return v.id
}
