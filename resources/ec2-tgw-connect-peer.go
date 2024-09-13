package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const EC2TGWConnectPeerResource = "EC2TGWConnectPeer"

func init() {
	registry.Register(&registry.Registration{
		Name:   EC2TGWConnectPeerResource,
		Scope:  nuke.Account,
		Lister: &EC2TGWConnectPeerLister{},
	})
}

type EC2TGWConnectPeerLister struct{}

func (l *EC2TGWConnectPeerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := ec2.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &ec2.DescribeTransitGatewayConnectPeersInput{}

	resp, err := svc.DescribeTransitGatewayConnectPeers(params)
	if err != nil {
		return nil, err
	}

	for _, connectPeer := range resp.TransitGatewayConnectPeers {
		resources = append(resources, &EC2TGWConnectPeer{
			svc:          svc,
			ID:           connectPeer.TransitGatewayConnectPeerId,
			State:        connectPeer.State,
			CreationTime: connectPeer.CreationTime,
			Tags:         connectPeer.Tags,
		})
	}

	return resources, nil
}

type EC2TGWConnectPeer struct {
	svc          *ec2.EC2
	ID           *string
	State        *string
	CreationTime *time.Time
	Tags         []*ec2.Tag
}

func (r *EC2TGWConnectPeer) Filter() error {
	if *r.State == "deleted" {
		return fmt.Errorf("already deleted")
	}
	return nil
}

func (r *EC2TGWConnectPeer) Remove(_ context.Context) error {
	_, err := r.svc.DeleteTransitGatewayConnectPeer(&ec2.DeleteTransitGatewayConnectPeerInput{
		TransitGatewayConnectPeerId: r.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *EC2TGWConnectPeer) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *EC2TGWConnectPeer) String() string {
	return *r.ID
}
