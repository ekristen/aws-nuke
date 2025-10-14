package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2ClientVpnEndpointAttachmentResource = "EC2ClientVpnEndpointAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2ClientVpnEndpointAttachmentResource,
		Scope:    nuke.Account,
		Resource: &EC2ClientVpnEndpointAttachments{},
		Lister:   &EC2ClientVpnEndpointAttachmentLister{},
	})
}

type EC2ClientVpnEndpointAttachmentLister struct{}

func (l *EC2ClientVpnEndpointAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	endpoints := make([]*string, 0)

	params := &ec2.DescribeClientVpnEndpointsInput{}
	err := svc.DescribeClientVpnEndpointsPages(params,
		func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
			for _, out := range page.ClientVpnEndpoints {
				endpoints = append(endpoints, out.ClientVpnEndpointId)
			}
			return true
		})
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, clientVpnEndpointID := range endpoints {
		params := &ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: clientVpnEndpointID,
		}
		err := svc.DescribeClientVpnTargetNetworksPages(params,
			func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
				for _, out := range page.ClientVpnTargetNetworks {
					resources = append(resources, &EC2ClientVpnEndpointAttachments{
						svc:                 svc,
						associationID:       out.AssociationId,
						clientVpnEndpointID: out.ClientVpnEndpointId,
						vpcID:               out.VpcId,
					})
				}
				return true
			})
		if err != nil {
			return nil, err
		}
	}

	return resources, nil
}

type EC2ClientVpnEndpointAttachments struct {
	svc                 *ec2.EC2
	associationID       *string
	clientVpnEndpointID *string
	vpcID               *string
}

func (e *EC2ClientVpnEndpointAttachments) Remove(_ context.Context) error {
	params := &ec2.DisassociateClientVpnTargetNetworkInput{
		AssociationId:       e.associationID,
		ClientVpnEndpointId: e.clientVpnEndpointID,
	}

	_, err := e.svc.DisassociateClientVpnTargetNetwork(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2ClientVpnEndpointAttachments) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(e.clientVpnEndpointID), ptr.ToString(e.vpcID))
}
