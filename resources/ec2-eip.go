package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2AddressResource = "EC2Address"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2AddressResource,
		Scope:    nuke.Account,
		Resource: &EC2Address{},
		Lister:   &EC2AddressLister{},
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
			svc:          svc,
			AllocationID: out.AllocationId,
			PublicIP:     out.PublicIp,
			Tags:         out.Tags,
		})
	}

	return resources, nil
}

type EC2Address struct {
	svc                *ec2.EC2
	AllocationID       *string
	PublicIP           *string
	NetworkBorderGroup *string
	Tags               []*ec2.Tag
}

func (r *EC2Address) Remove(_ context.Context) error {
	_, err := r.svc.ReleaseAddress(&ec2.ReleaseAddressInput{
		AllocationId:       r.AllocationID,
		NetworkBorderGroup: r.NetworkBorderGroup,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2Address) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2Address) String() string {
	return *r.PublicIP
}
