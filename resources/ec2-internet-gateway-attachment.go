package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2InternetGatewayAttachmentResource = "EC2InternetGatewayAttachment"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2InternetGatewayAttachmentResource,
		Scope:    nuke.Account,
		Resource: &EC2InternetGatewayAttachment{},
		Lister:   &EC2InternetGatewayAttachmentLister{},
		DeprecatedAliases: []string{
			"EC2InternetGatewayAttachement",
		},
	})
}

type EC2InternetGatewayAttachmentLister struct{}

func (l *EC2InternetGatewayAttachmentLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, vpc := range resp.Vpcs {
		params := &ec2.DescribeInternetGatewaysInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("attachment.vpc-id"),
					Values: []*string{vpc.VpcId},
				},
			},
		}

		resp, err := svc.DescribeInternetGateways(params)
		if err != nil {
			return nil, err
		}

		for _, igw := range resp.InternetGateways {
			resources = append(resources, &EC2InternetGatewayAttachment{
				svc:        svc,
				vpcID:      vpc.VpcId,
				vpcOwnerID: vpc.OwnerId,
				vpcTags:    vpc.Tags,
				igwID:      igw.InternetGatewayId,
				igwOwnerID: igw.OwnerId,
				igwTags:    igw.Tags,
				defaultVPC: *vpc.IsDefault,
			})
		}
	}

	return resources, nil
}

type EC2InternetGatewayAttachment struct {
	svc        *ec2.EC2
	accountID  *string
	vpcID      *string
	vpcOwnerID *string
	vpcTags    []*ec2.Tag
	igwID      *string
	igwOwnerID *string
	igwTags    []*ec2.Tag
	defaultVPC bool
}

func (r *EC2InternetGatewayAttachment) Filter() error {
	if ptr.ToString(r.igwOwnerID) != ptr.ToString(r.accountID) {
		return errors.New("not owned by account, likely shared")
	}

	return nil
}

func (r *EC2InternetGatewayAttachment) Remove(_ context.Context) error {
	params := &ec2.DetachInternetGatewayInput{
		VpcId:             r.vpcID,
		InternetGatewayId: r.igwID,
	}

	_, err := r.svc.DetachInternetGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2InternetGatewayAttachment) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DefaultVPC", r.defaultVPC)
	properties.SetWithPrefix("vpc", "OwnerID", r.vpcOwnerID)
	properties.SetWithPrefix("igw", "OwnerID", r.igwOwnerID)

	for _, tagValue := range r.igwTags {
		properties.SetTagWithPrefix("igw", tagValue.Key, tagValue.Value)
	}
	for _, tagValue := range r.vpcTags {
		properties.SetTagWithPrefix("vpc", tagValue.Key, tagValue.Value)
	}

	return properties
}

func (r *EC2InternetGatewayAttachment) String() string {
	return fmt.Sprintf("%s -> %s", ptr.ToString(r.igwID), ptr.ToString(r.vpcID))
}
