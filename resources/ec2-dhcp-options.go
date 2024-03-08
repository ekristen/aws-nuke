package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2DHCPOptionResource = "EC2DHCPOption"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2DHCPOptionResource,
		Scope:  nuke.Account,
		Lister: &EC2DHCPOptionLister{},
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
	id         *string
	tags       []*ec2.Tag
	defaultVPC bool
	ownerID    *string
}

func (e *EC2DHCPOption) Remove(_ context.Context) error {
	params := &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: e.id,
	}

	_, err := e.svc.DeleteDhcpOptions(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2DHCPOption) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DefaultVPC", e.defaultVPC)
	properties.Set("OwnerID", e.ownerID)

	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *EC2DHCPOption) String() string {
	return ptr.ToString(e.id)
}
