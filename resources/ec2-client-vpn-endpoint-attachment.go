package resources

import (
	"context"

	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2ClientVpnEndpointAttachmentResource = "EC2ClientVpnEndpointAttachment"

func init() {
	resource.Register(&resource.Registration{
		Name:   EC2ClientVpnEndpointAttachmentResource,
		Scope:  nuke.Account,
		Lister: &EC2ClientVpnEndpointAttachmentLister{},
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
	for _, clientVpnEndpointId := range endpoints {
		params := &ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: clientVpnEndpointId,
		}
		err := svc.DescribeClientVpnTargetNetworksPages(params,
			func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
				for _, out := range page.ClientVpnTargetNetworks {
					resources = append(resources, &EC2ClientVpnEndpointAttachments{
						svc:                 svc,
						associationId:       out.AssociationId,
						clientVpnEndpointId: out.ClientVpnEndpointId,
						vpcId:               out.VpcId,
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
	associationId       *string
	clientVpnEndpointId *string
	vpcId               *string
}

func (e *EC2ClientVpnEndpointAttachments) Remove(_ context.Context) error {
	params := &ec2.DisassociateClientVpnTargetNetworkInput{
		AssociationId:       e.associationId,
		ClientVpnEndpointId: e.clientVpnEndpointId,
	}

	_, err := e.svc.DisassociateClientVpnTargetNetwork(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2ClientVpnEndpointAttachments) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(e.clientVpnEndpointId), ptr.ToString(e.vpcId))
}
