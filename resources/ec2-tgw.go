package resources

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2TGWResource = "EC2TGW"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2TGWResource,
		Scope:  nuke.Account,
		Lister: &EC2TGWLister{},
		DependsOn: []string{
			EC2TGWAttachmentResource,
		},
	})
}

type EC2TGWLister struct{}

func (l *EC2TGWLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	params := &ec2.DescribeTransitGatewaysInput{}
	resources := make([]resource.Resource, 0)
	for {
		resp, err := svc.DescribeTransitGateways(params)
		if err != nil {
			return nil, err
		}

		for _, tgw := range resp.TransitGateways {
			resources = append(resources, &EC2TGW{
				svc: svc,
				tgw: tgw,
			})
		}

		if resp.NextToken == nil {
			break
		}

		params = &ec2.DescribeTransitGatewaysInput{
			NextToken: resp.NextToken,
		}
	}

	return resources, nil
}

type EC2TGW struct {
	svc *ec2.EC2
	tgw *ec2.TransitGateway
}

func (e *EC2TGW) Remove(_ context.Context) error {
	params := &ec2.DeleteTransitGatewayInput{
		TransitGatewayId: e.tgw.TransitGatewayId,
	}

	_, err := e.svc.DeleteTransitGateway(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2TGW) Filter() error {
	if *e.tgw.State == awsutil.StateDeleted {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (e *EC2TGW) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range e.tgw.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	properties.
		Set("ID", e.tgw.TransitGatewayId).
		Set("OwnerId", e.tgw.OwnerId)

	return properties
}

func (e *EC2TGW) String() string {
	return *e.tgw.TransitGatewayId
}
