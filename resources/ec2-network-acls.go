package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2NetworkACLResource = "EC2NetworkACL"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2NetworkACLResource,
		Scope:    nuke.Account,
		Resource: &EC2NetworkACL{},
		Lister:   &EC2NetworkACLLister{},
	})
}

type EC2NetworkACLLister struct{}

func (l *EC2NetworkACLLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeNetworkAcls(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.NetworkAcls {
		resources = append(resources, &EC2NetworkACL{
			svc:       svc,
			id:        out.NetworkAclId,
			isDefault: out.IsDefault,
			tags:      out.Tags,
			ownerID:   out.OwnerId,
		})
	}

	return resources, nil
}

type EC2NetworkACL struct {
	svc       *ec2.EC2
	id        *string
	isDefault *bool
	tags      []*ec2.Tag
	ownerID   *string
}

func (e *EC2NetworkACL) Filter() error {
	if ptr.ToBool(e.isDefault) {
		return fmt.Errorf("cannot delete default VPC")
	}

	return nil
}

func (e *EC2NetworkACL) Remove(_ context.Context) error {
	params := &ec2.DeleteNetworkAclInput{
		NetworkAclId: e.id,
	}

	_, err := e.svc.DeleteNetworkAcl(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2NetworkACL) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}
	properties.Set("ID", e.id)
	properties.Set("OwnerID", e.ownerID)
	return properties
}

func (e *EC2NetworkACL) String() string {
	return ptr.ToString(e.id)
}
