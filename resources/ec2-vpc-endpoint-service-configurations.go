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

const EC2VPCEndpointServiceConfigurationResource = "EC2VPCEndpointServiceConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2VPCEndpointServiceConfigurationResource,
		Scope:  nuke.Account,
		Lister: &EC2VPCEndpointServiceConfigurationLister{},
	})
}

type EC2VPCEndpointServiceConfigurationLister struct{}

func (l *EC2VPCEndpointServiceConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.DescribeVpcEndpointServiceConfigurations(params)
		if err != nil {
			return nil, err
		}

		for _, serviceConfig := range resp.ServiceConfigurations {
			resources = append(resources, &EC2VPCEndpointServiceConfiguration{
				svc:  svc,
				id:   serviceConfig.ServiceId,
				name: serviceConfig.ServiceName,
				tags: serviceConfig.Tags,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2VPCEndpointServiceConfiguration struct {
	svc  *ec2.EC2
	id   *string
	name *string
	tags []*ec2.Tag
}

func (e *EC2VPCEndpointServiceConfiguration) Remove(_ context.Context) error {
	params := &ec2.DeleteVpcEndpointServiceConfigurationsInput{
		ServiceIds: []*string{e.id},
	}

	_, err := e.svc.DeleteVpcEndpointServiceConfigurations(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2VPCEndpointServiceConfiguration) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", e.id)
	properties.Set("Name", e.name)

	for _, tag := range e.tags {
		properties.SetTag(tag.Key, tag.Value)
	}

	return properties
}

func (e *EC2VPCEndpointServiceConfiguration) String() string {
	return *e.id
}
