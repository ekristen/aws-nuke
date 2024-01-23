package resources

import (
	"context"

	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/nuke"
)

const EC2VPNConnectionResource = "EC2VPNConnection"

func init() {
	resource.Register(&resource.Registration{
		Name:   EC2VPNConnectionResource,
		Scope:  nuke.Account,
		Lister: &EC2VPNConnectionLister{},
		DeprecatedAliases: []string{
			"EC2VpnConnection",
		},
	})
}

type EC2VPNConnectionLister struct{}

func (l *EC2VPNConnectionLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := ec2.New(opts.Session)

	params := &ec2.DescribeVpnConnectionsInput{}
	resp, err := svc.DescribeVpnConnections(params)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0)
	for _, out := range resp.VpnConnections {
		resources = append(resources, &EC2VPNConnection{
			svc:  svc,
			conn: out,
		})
	}

	return resources, nil
}

type EC2VPNConnection struct {
	svc  *ec2.EC2
	conn *ec2.VpnConnection
}

func (v *EC2VPNConnection) Filter() error {
	if *v.conn.State == "deleted" {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (v *EC2VPNConnection) Remove(_ context.Context) error {
	params := &ec2.DeleteVpnConnectionInput{
		VpnConnectionId: v.conn.VpnConnectionId,
	}

	_, err := v.svc.DeleteVpnConnection(params)
	if err != nil {
		return err
	}

	return nil
}

func (v *EC2VPNConnection) Properties() types.Properties {
	properties := types.NewProperties()
	for _, tagValue := range v.conn.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}
	return properties
}

func (v *EC2VPNConnection) String() string {
	return *v.conn.VpnConnectionId
}
