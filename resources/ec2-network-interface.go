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

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2NetworkInterfaceResource = "EC2NetworkInterface"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2NetworkInterfaceResource,
		Scope:    nuke.Account,
		Resource: &EC2NetworkInterface{},
		Lister:   &EC2NetworkInterfaceLister{},
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

func (r *EC2NetworkInterface) Remove(_ context.Context) error {
	if r.eni.Attachment != nil {
		_, err := r.svc.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: r.eni.Attachment.AttachmentId,
			Force:        aws.Bool(true),
		})
		if err != nil {
			if r.eni.Attachment.AttachmentId != nil {
				expected := fmt.Sprintf("The interface attachment '%s' does not exist.", *r.eni.Attachment.AttachmentId)
				if !strings.Contains(err.Error(), expected) {
					return err
				}
			}
		}
	}

	params := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: r.eni.NetworkInterfaceId,
	}

	_, err := r.svc.DeleteNetworkInterface(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2NetworkInterface) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range r.eni.TagSet {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.
		Set("ID", r.eni.NetworkInterfaceId).
		Set("VPC", r.eni.VpcId).
		Set("AvailabilityZone", r.eni.AvailabilityZone).
		Set("PrivateIPAddress", r.eni.PrivateIpAddress).
		Set("SubnetID", r.eni.SubnetId).
		Set("Status", r.eni.Status)
	return properties
}

func (r *EC2NetworkInterface) String() string {
	return *r.eni.NetworkInterfaceId
}
