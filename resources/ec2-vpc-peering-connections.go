package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2VPCPeeringConnectionResource = "EC2VPCPeeringConnection"

func init() {
	resource.Register(&resource.Registration{
		Name:   EC2VPCPeeringConnectionResource,
		Scope:  nuke.Account,
		Lister: &EC2VPCPeeringConnectionLister{},
	})
}

type EC2VPCPeeringConnectionLister struct{}

func (l *EC2VPCPeeringConnectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	// filter should be set as deleted vpc connections are returned
	params := &ec2.DescribeVpcPeeringConnectionsInput{}

	resp, err := svc.DescribeVpcPeeringConnections(params)
	if err != nil {
		return nil, err
	}

	for _, peeringConfig := range resp.VpcPeeringConnections {
		resources = append(resources, &EC2VPCPeeringConnection{
			svc:    svc,
			id:     peeringConfig.VpcPeeringConnectionId,
			status: peeringConfig.Status.Code,
		})
	}

	return resources, nil
}

type EC2VPCPeeringConnection struct {
	svc    *ec2.EC2
	id     *string
	status *string
}

func (p *EC2VPCPeeringConnection) Filter() error {
	if *p.status == "deleting" || *p.status == "deleted" {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (p *EC2VPCPeeringConnection) Remove(_ context.Context) error {
	params := &ec2.DeleteVpcPeeringConnectionInput{
		VpcPeeringConnectionId: p.id,
	}

	_, err := p.svc.DeleteVpcPeeringConnection(params)
	if err != nil {
		return err
	}
	return nil
}

func (p *EC2VPCPeeringConnection) String() string {
	return *p.id
}
