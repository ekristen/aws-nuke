package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2KeyPairResource = "EC2KeyPair"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2KeyPairResource,
		Scope:  nuke.Account,
		Lister: &EC2KeyPairLister{},
	})
}

type EC2KeyPairLister struct{}

func (l *EC2KeyPairLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeKeyPairs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.KeyPairs {
		resources = append(resources, &EC2KeyPair{
			svc:  svc,
			name: *out.KeyName,
			tags: out.Tags,
		})
	}

	return resources, nil
}

type EC2KeyPair struct {
	svc  *ec2.EC2
	name string
	tags []*ec2.Tag
}

func (e *EC2KeyPair) Remove(_ context.Context) error {
	params := &ec2.DeleteKeyPairInput{
		KeyName: &e.name,
	}

	_, err := e.svc.DeleteKeyPair(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2KeyPair) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", e.name)

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (e *EC2KeyPair) String() string {
	return e.name
}
