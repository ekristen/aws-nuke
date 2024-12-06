package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gotidy/ptr"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2InternetGatewayResource = "EC2InternetGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2InternetGatewayResource,
		Scope:    nuke.Account,
		Resource: &EC2InternetGateway{},
		Lister:   &EC2InternetGatewayLister{},
	})
}

type EC2InternetGatewayLister struct{}

func (l *EC2InternetGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeInternetGateways(nil)
	if err != nil {
		return nil, err
	}

	defVpcID := ""
	if defVpc := DefaultVpc(svc); defVpc != nil {
		defVpcID = *defVpc.VpcId
	}

	resources := make([]resource.Resource, 0)
	for _, igw := range resp.InternetGateways {
		resources = append(resources, &EC2InternetGateway{
			svc:        svc,
			igw:        igw,
			defaultVPC: HasVpcAttachment(&defVpcID, igw.Attachments),
		})
	}

	return resources, nil
}

func HasVpcAttachment(vpcID *string, attachments []*ec2.InternetGatewayAttachment) bool {
	if *vpcID == "" {
		return false
	}

	for _, attach := range attachments {
		if *vpcID == *attach.VpcId {
			return true
		}
	}
	return false
}

type EC2InternetGateway struct {
	svc        *ec2.EC2
	igw        *ec2.InternetGateway
	defaultVPC bool
}

func (e *EC2InternetGateway) Remove(_ context.Context) error {
	params := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: e.igw.InternetGatewayId,
	}

	_, err := e.svc.DeleteInternetGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2InternetGateway) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.igw.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("DefaultVPC", e.defaultVPC)
	properties.Set("OwnerID", e.igw.OwnerId)
	return properties
}

func (e *EC2InternetGateway) String() string {
	return ptr.ToString(e.igw.InternetGatewayId)
}
