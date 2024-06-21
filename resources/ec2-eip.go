package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2AddressResource = "EC2Address"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2AddressResource,
		Scope:  nuke.Account,
		Lister: &EC2AddressLister{},
	})
}

type EC2AddressLister struct{}

func (l *EC2AddressLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeAddressesInput{}
	resp, err := svc.DescribeAddresses(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.Addresses {
		resources = append(resources, &EC2Address{
			svc: svc,
			eip: out,
			id:  ptr.ToString(out.AllocationId),
			ip:  ptr.ToString(out.PublicIp),
		})
	}

	return resources, nil
}

type EC2Address struct {
	svc *ec2.EC2
	eip *ec2.Address
	id  string
	ip  string
}

func (e *EC2Address) Remove(_ context.Context) error {
	_, err := e.svc.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId: &e.id,
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2Address) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.eip.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.Set("AllocationID", e.id)
	return properties
}

func (e *EC2Address) String() string {
	return e.ip
}
