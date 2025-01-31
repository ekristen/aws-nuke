package resources

import (
	"context"
	"errors"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2DHCPOptionResource = "EC2DHCPOption"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2DHCPOptionResource,
		Scope:    nuke.Account,
		Resource: &EC2DHCPOption{},
		Lister:   &EC2DHCPOptionLister{},
		DeprecatedAliases: []string{
			"EC2DhcpOptions",
			"EC2DHCPOptions",
		},
	})
}

type EC2DHCPOptionLister struct{}

func (l *EC2DHCPOptionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeDhcpOptions(&ec2.DescribeDhcpOptionsInput{})
	if err != nil {
		return nil, err
	}

	defVpcDhcpOptsID := ""
	if defVpc := DefaultVpc(svc); defVpc != nil {
		defVpcDhcpOptsID = ptr.ToString(defVpc.DhcpOptionsId)
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.DhcpOptions {
		resources = append(resources, &EC2DHCPOption{
			svc:        svc,
			accountID:  opts.AccountID,
			id:         out.DhcpOptionsId,
			tags:       out.Tags,
			defaultVPC: defVpcDhcpOptsID == ptr.ToString(out.DhcpOptionsId),
			ownerID:    out.OwnerId,
		})
	}

	return resources, nil
}

type EC2DHCPOption struct {
	svc        *ec2.EC2
	accountID  *string
	id         *string
	tags       []*ec2.Tag
	defaultVPC bool
	ownerID    *string
}

func (r *EC2DHCPOption) Filter() error {
	if ptr.ToString(r.ownerID) != ptr.ToString(r.accountID) {
		return errors.New("not owned by account, likely shared")
	}

	return nil
}

func (r *EC2DHCPOption) Remove(_ context.Context) error {
	params := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: r.id,
	}

	_, err := r.svc.DeleteDhcpOptions(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2DHCPOption) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DefaultVPC", r.defaultVPC)
	properties.Set("OwnerID", r.ownerID)

	for _, tagValue := range r.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (r *EC2DHCPOption) String() string {
	return ptr.ToString(r.id)
}
