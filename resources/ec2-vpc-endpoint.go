package resources

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2VPCEndpointResource = "EC2VPCEndpoint"

func init() {
	registry.Register(&registry.Registration{
		Name:     EC2VPCEndpointResource,
		Scope:    nuke.Account,
		Resource: &EC2VPCEndpoint{},
		Lister:   &EC2VPCEndpointLister{},
		DependsOn: []string{
			EC2VPCEndpointConnectionResource,
			EC2VPCEndpointServiceConfigurationResource,
		},
		DeprecatedAliases: []string{
			"EC2VpcEndpoint",
		},
	})
}

type EC2VPCEndpointLister struct{}

func (l *EC2VPCEndpointLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeVpcs(nil)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, vpc := range resp.Vpcs {
		params := &ec2.DescribeVpcEndpointsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []*string{vpc.VpcId},
				},
			},
		}

		resp, err := svc.DescribeVpcEndpoints(params)
		if err != nil {
			return nil, err
		}

		for _, vpcEndpoint := range resp.VpcEndpoints {
			resources = append(resources, &EC2VPCEndpoint{
				svc:  svc,
				id:   vpcEndpoint.VpcEndpointId,
				tags: vpcEndpoint.Tags,
			})
		}
	}

	return resources, nil
}

type EC2VPCEndpoint struct {
	svc  *ec2.EC2
	id   *string
	tags []*ec2.Tag
}

func (e *EC2VPCEndpoint) Remove(_ context.Context) error {
	params := &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: []*string{e.id},
	}

	_, err := e.svc.DeleteVpcEndpoints(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2VPCEndpoint) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *EC2VPCEndpoint) String() string {
	return *e.id
}
