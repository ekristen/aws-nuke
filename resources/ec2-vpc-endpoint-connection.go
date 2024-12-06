package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VPCEndpointConnectionResource = "EC2VPCEndpointConnection" //nolint:gosec,nolintlint

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VPCEndpointConnectionResource,
		Scope:    nuke.Account,
		Resource: &EC2VPCEndpointConnection{},
		Lister:   &EC2VPCEndpointConnectionLister{},
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
				ServiceID:     endpointConnection.ServiceId,
				VPCEndpointID: endpointConnection.VpcEndpointId,
				State:         endpointConnection.VpcEndpointState,
				Owner:         endpointConnection.VpcEndpointOwner,
				Tags:          endpointConnection.Tags,
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
	ServiceID     *string
	VPCEndpointID *string
	State         *string
	Owner         *string
	Tags          []*ec2.Tag
}

func (r *EC2VPCEndpointConnection) Filter() error {
	if *r.State == awsutil.StateDeleting || *r.State == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}

	if strings.EqualFold(ptr.ToString(r.State), awsutil.StateRejected) {
		return fmt.Errorf("non-deletable state: rejected")
	}

	return nil
}

func (r *EC2VPCEndpointConnection) Remove(_ context.Context) error {
	params := &ec2.RejectVpcEndpointConnectionsInput{
		ServiceId: r.ServiceID,
		VpcEndpointIds: []*string{
			r.VPCEndpointID,
		},
	}

	_, err := r.svc.RejectVpcEndpointConnections(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2VPCEndpointConnection) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r).
		Set("VpcEndpointID", r.VPCEndpointID) // TODO(v4): remove the extra set
}

func (r *EC2VPCEndpointConnection) String() string {
	return *r.ServiceID
}
