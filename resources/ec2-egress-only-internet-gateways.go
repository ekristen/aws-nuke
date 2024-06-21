package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2EgressOnlyInternetGatewayResource = "EC2EgressOnlyInternetGateway"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2EgressOnlyInternetGatewayResource,
		Scope:  nuke.Account,
		Lister: &EC2EgressOnlyInternetGatewayLister{},
	})
}

type EC2EgressOnlyInternetGatewayLister struct{}

func (l *EC2EgressOnlyInternetGatewayLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)
	igwInputParams := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		MaxResults: aws.Int64(255),
	}

	for {
		resp, err := svc.DescribeEgressOnlyInternetGateways(igwInputParams)
		if err != nil {
			return nil, err
		}

		for _, igw := range resp.EgressOnlyInternetGateways {
			resources = append(resources, &EC2EgressOnlyInternetGateway{
				svc: svc,
				igw: igw,
			})
		}

		if resp.NextToken == nil {
			break
		}

		igwInputParams.NextToken = resp.NextToken
	}

	return resources, nil
}

type EC2EgressOnlyInternetGateway struct {
	svc *ec2.EC2
	igw *ec2.EgressOnlyInternetGateway
}

func (e *EC2EgressOnlyInternetGateway) Remove(_ context.Context) error {
	params := &ec2.DeleteEgressOnlyInternetGatewayInput{
		EgressOnlyInternetGatewayId: e.igw.EgressOnlyInternetGatewayId,
	}

	_, err := e.svc.DeleteEgressOnlyInternetGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2EgressOnlyInternetGateway) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.igw.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (e *EC2EgressOnlyInternetGateway) String() string {
	return ptr.ToString(e.igw.EgressOnlyInternetGatewayId)
}
