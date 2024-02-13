package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2SubnetResource = "EC2Subnet"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2SubnetResource,
		Scope:  nuke.Account,
		Lister: &EC2SubnetLister{},
		DependsOn: []string{
			EC2NetworkInterfaceResource,
		},
	})
}

type EC2Subnet struct {
	svc        *ec2.EC2
	subnet     *ec2.Subnet
	defaultVPC bool
}

type EC2SubnetLister struct{}

func (l *EC2SubnetLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeSubnetsInput{}
	resp, err := svc.DescribeSubnets(params)
	if err != nil {
		return nil, err
	}

	defVpcId := ""
	if defVpc := DefaultVpc(svc); defVpc != nil {
		defVpcId = *defVpc.VpcId
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.Subnets {
		resources = append(resources, &EC2Subnet{
			svc:        svc,
			subnet:     out,
			defaultVPC: defVpcId == *out.VpcId,
		})
	}

	return resources, nil
}

func (e *EC2Subnet) Remove(_ context.Context) error {
	params := &ec2.DeleteSubnetInput{
		SubnetId: e.subnet.SubnetId,
	}

	_, err := e.svc.DeleteSubnet(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2Subnet) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DefaultForAz", e.subnet.DefaultForAz)
	properties.Set("DefaultVPC", e.defaultVPC)
	properties.Set("OwnerID", e.subnet.OwnerId)
	properties.Set("VpcID", e.subnet.VpcId)

	for _, tagValue := range e.subnet.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *EC2Subnet) String() string {
	return *e.subnet.SubnetId
}
