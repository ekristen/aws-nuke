package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2ClientVpnEndpointResource = "EC2ClientVpnEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2ClientVpnEndpointResource,
		Scope:    nuke.Account,
		Resource: &EC2ClientVpnEndpoint{},
		Lister:   &EC2ClientVpnEndpointLister{},
		DependsOn: []string{
			EC2ClientVpnEndpointAttachmentResource,
		},
	})
}

type EC2ClientVpnEndpointLister struct{}

func (l *EC2ClientVpnEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)
	params := &ec2.DescribeClientVpnEndpointsInput{}

	err := svc.DescribeClientVpnEndpointsPages(params,
		func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
			for _, out := range page.ClientVpnEndpoints {
				resources = append(resources, &EC2ClientVpnEndpoint{
					svc: svc,
					id:  *out.ClientVpnEndpointId,
				})
			}

			return true
		})
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type EC2ClientVpnEndpoint struct {
	svc     *ec2.EC2
	id      string
	cveTags []*ec2.Tag
}

func (c *EC2ClientVpnEndpoint) Remove(_ context.Context) error {
	params := &ec2.DeleteClientVpnEndpointInput{
		ClientVpnEndpointId: &c.id,
	}

	_, err := c.svc.DeleteClientVpnEndpoint(params)
	if err != nil {
		return err
	}

	return nil
}

func (c *EC2ClientVpnEndpoint) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range c.cveTags {
		properties.SetTagWithPrefix("cve", tagValue.Key, tagValue.Value)
	}
	return properties
}

func (c *EC2ClientVpnEndpoint) String() string {
	return c.id
}
