package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2CustomerGatewayResource = "EC2CustomerGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2CustomerGatewayResource,
		Scope:  nuke.Account,
		Lister: &EC2CustomerGatewayLister{},
	})
}

type EC2CustomerGatewayLister struct{}

func (l *EC2CustomerGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)

	params := &ec2.DescribeCustomerGatewaysInput{}
	resp, err := svc.DescribeCustomerGateways(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.CustomerGateways {
		resources = append(resources, &EC2CustomerGateway{
			svc:   svc,
			id:    *out.CustomerGatewayId,
			state: *out.State,
		})
	}

	return resources, nil
}

type EC2CustomerGateway struct {
	svc   *ec2.EC2
	id    string
	state string
}

func (c *EC2CustomerGateway) Filter() error {
	if c.state == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (c *EC2CustomerGateway) Remove(_ context.Context) error {
	params := &ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: &c.id,
	}

	_, err := c.svc.DeleteCustomerGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *EC2CustomerGateway) String() string {
	return c.id
}
