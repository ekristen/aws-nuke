package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2KeyPairResource = "EC2KeyPair"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2KeyPairResource,
		Scope:    nuke.Account,
		Resource: &EC2KeyPair{},
		Lister:   &EC2KeyPairLister{},
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
			svc:        svc,
			Name:       out.KeyName,
			Tags:       out.Tags,
			KeyType:    out.KeyType,
			CreateTime: out.CreateTime,
		})
	}

	return resources, nil
}

type EC2KeyPair struct {
	svc        *ec2.EC2
	Name       *string
	Tags       []*ec2.Tag
	KeyType    *string
	CreateTime *time.Time
}

func (r *EC2KeyPair) Remove(_ context.Context) error {
	params := &ec2.DeleteKeyPairInput{
		KeyName: r.Name,
	}

	_, err := r.svc.DeleteKeyPair(params)
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2KeyPair) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2KeyPair) String() string {
	return *r.Name
}
