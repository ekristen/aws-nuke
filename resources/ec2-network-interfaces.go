package resources

import (
	"context"

	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2NetworkInterfaceResource = "EC2NetworkInterface"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2NetworkInterfaceResource,
		Scope:  nuke.Account,
		Lister: &EC2NetworkInterfaceLister{},
	})
}

type EC2NetworkInterfaceLister struct{}

func (l *EC2NetworkInterfaceLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeNetworkInterfaces(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.NetworkInterfaces {

		resources = append(resources, &EC2NetworkInterface{
			svc: svc,
			eni: out,
		})
	}

	return resources, nil
}

type EC2NetworkInterface struct {
	svc *ec2.EC2
	eni *ec2.NetworkInterface
}

func (e *EC2NetworkInterface) Remove(_ context.Context) error {
	if e.eni.Attachment != nil {
		_, err := e.svc.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: e.eni.Attachment.AttachmentId,
			Force:        aws.Bool(true),
		})
		if err != nil {
			if e.eni.Attachment.AttachmentId != nil {
				expected := fmt.Sprintf("The interface attachment '%s' does not exist.", *e.eni.Attachment.AttachmentId)
				if !strings.Contains(err.Error(), expected) {
					return err
				}
			}

		}
	}

	params := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: e.eni.NetworkInterfaceId,
	}

	_, err := e.svc.DeleteNetworkInterface(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2NetworkInterface) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range e.eni.TagSet {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.
		Set("ID", e.eni.NetworkInterfaceId).
		Set("VPC", e.eni.VpcId).
		Set("AvailabilityZone", e.eni.AvailabilityZone).
		Set("PrivateIPAddress", e.eni.PrivateIpAddress).
		Set("SubnetID", e.eni.SubnetId).
		Set("Status", e.eni.Status)
	return properties
}

func (e *EC2NetworkInterface) String() string {
	return *e.eni.NetworkInterfaceId
}
