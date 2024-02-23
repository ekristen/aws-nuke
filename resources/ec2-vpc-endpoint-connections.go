package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2VPCEndpointConnectionResource = "EC2VPCEndpointConnection" //nolint:gosec,nolintlint

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2VPCEndpointConnectionResource,
		Scope:  nuke.Account,
		Lister: &EC2VPCEndpointConnectionLister{},
	})
}

type EC2VPCEndpointConnectionLister struct{}

func (l *EC2VPCEndpointConnectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &ec2.DescribeVpcEndpointConnectionsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.DescribeVpcEndpointConnections(params)
		if err != nil {
			return nil, err
		}

		for _, endpointConnection := range resp.VpcEndpointConnections {
			resources = append(resources, &EC2VPCEndpointConnection{
				svc:           svc,
				vpcEndpointID: endpointConnection.VpcEndpointId,
				serviceID:     endpointConnection.ServiceId,
				state:         endpointConnection.VpcEndpointState,
				owner:         endpointConnection.VpcEndpointOwner,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2VPCEndpointConnection struct {
	svc           *ec2.EC2
	serviceID     *string
	vpcEndpointID *string
	state         *string
	owner         *string
}

func (c *EC2VPCEndpointConnection) Filter() error {
	if *c.state == awsutil.StateDeleting || *c.state == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (c *EC2VPCEndpointConnection) Remove(_ context.Context) error {
	params := &ec2.RejectVpcEndpointConnectionsInput{
		ServiceId: c.serviceID,
		VpcEndpointIds: []*string{
			c.vpcEndpointID,
		},
	}

	_, err := c.svc.RejectVpcEndpointConnections(params)
	if err != nil {
		return err
	}
	return nil
}

func (c *EC2VPCEndpointConnection) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("VpcEndpointID", c.vpcEndpointID)
	properties.Set("State", c.state)
	properties.Set("Owner", c.owner)
	return properties
}

func (c *EC2VPCEndpointConnection) String() string {
	return *c.serviceID
}
