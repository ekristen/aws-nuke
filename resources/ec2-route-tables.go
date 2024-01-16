package resources

import (
	"context"

	"fmt"
	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2RouteTableResource = "EC2RouteTable"

func init() {
	resource.Register(resource.Registration{
		Name:   EC2RouteTableResource,
		Scope:  nuke.Account,
		Lister: &EC2RouteTableLister{},
		DependsOn: []string{
			EC2SubnetResource,
		},
	})
}

type EC2RouteTable struct {
	svc        *ec2.EC2
	routeTable *ec2.RouteTable
	defaultVPC bool
	vpc        *ec2.Vpc
}

type EC2RouteTableLister struct{}

func (l *EC2RouteTableLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)

	resp, err := svc.DescribeRouteTables(nil)
	if err != nil {
		return nil, err
	}

	defVpcId := ""
	if defVpc := DefaultVpc(svc); defVpc != nil {
		defVpcId = ptr.ToString(defVpc.VpcId)
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.RouteTables {
		vpc, err := GetVPC(svc, out.VpcId)
		if err != nil {
			return resources, nil
		}

		resources = append(resources, &EC2RouteTable{
			svc:        svc,
			routeTable: out,
			defaultVPC: defVpcId == ptr.ToString(out.VpcId),
			vpc:        vpc,
		})
	}

	return resources, nil
}

func (e *EC2RouteTable) Filter() error {
	for _, association := range e.routeTable.Associations {
		if *association.Main {
			return fmt.Errorf("main route tables cannot be deleted")
		}
	}

	return nil
}

func (e *EC2RouteTable) Remove(_ context.Context) error {
	params := &ec2.DeleteRouteTableInput{
		RouteTableId: e.routeTable.RouteTableId,
	}

	_, err := e.svc.DeleteRouteTable(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *EC2RouteTable) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("DefaultVPC", e.defaultVPC)
	properties.Set("vpcID", e.routeTable.VpcId)

	for _, tagValue := range e.routeTable.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	for _, tagValue := range e.vpc.Tags {
		properties.SetTagWithPrefix("vpc", tagValue.Key, tagValue.Value)
	}

	return properties
}

func (e *EC2RouteTable) String() string {
	return ptr.ToString(e.routeTable.RouteTableId)
}
