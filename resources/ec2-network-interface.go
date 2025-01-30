package resources

import (
	"context"
	"errors"
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

	resources := make([]resource.Resource, 0)

	params := &ec2.DescribeNetworkInterfacesInput{
		MaxResults: aws.Int64(1000),
	}

	for {
		resp, err := svc.DescribeNetworkInterfaces(params)
		if err != nil {
			return nil, err
		}

		for _, out := range resp.NetworkInterfaces {
			newResource := &EC2NetworkInterface{
				svc:              svc,
				accountID:        opts.AccountID,
				ID:               out.NetworkInterfaceId,
				VPC:              out.VpcId,
				AvailabilityZone: out.AvailabilityZone,
				PrivateIPAddress: out.PrivateIpAddress,
				SubnetID:         out.SubnetId,
				Status:           out.Status,
				OwnerID:          out.OwnerId,
			}

			if out.Attachment != nil {
				newResource.AttachmentID = out.Attachment.AttachmentId
			}

			resources = append(resources, newResource)
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2NetworkInterface struct {
	svc       *ec2.EC2
	accountID *string

	ID               *string
	VPC              *string
	AttachmentID     *string
	AvailabilityZone *string
	PrivateIPAddress *string
	SubnetID         *string
	Status           *string
	OwnerID          *string
	Tags             []*ec2.Tag
}

func (r *EC2NetworkInterface) Filter() error {
	if *r.OwnerID != *r.accountID {
		return errors.New("not owned by account, likely RAM shared")
	}

	return nil
}

func (r *EC2NetworkInterface) Remove(_ context.Context) error {
	if r.AttachmentID != nil {
		_, err := r.svc.DetachNetworkInterface(&ec2.DetachNetworkInterfaceInput{
			AttachmentId: r.AttachmentID,
			Force:        aws.Bool(true),
		})
		if err != nil {
			if r.AttachmentID != nil {
				expected := fmt.Sprintf("The interface attachment '%s' does not exist.", *r.AttachmentID)
				if !strings.Contains(err.Error(), expected) {
					return err
				}
			}
		}
	}

	params := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: r.ID,
	}

	_, err := r.svc.DeleteNetworkInterface(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2NetworkInterface) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2NetworkInterface) String() string {
	return *r.ID
}
