package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/service/networkmanager" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NetworkManagerConnectPeerResource = "NetworkManagerConnectPeer"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkManagerConnectPeerResource,
		Scope:    nuke.Account,
		Resource: &NetworkManagerConnectPeer{},
		Lister:   &NetworkManagerConnectPeerLister{},
	})
}

type NetworkManagerConnectPeerLister struct{}

func (l *NetworkManagerConnectPeerLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := networkmanager.New(opts.Session)
	params := &networkmanager.ListConnectPeersInput{}
	resources := make([]resource.Resource, 0)

	resp, err := svc.ListConnectPeers(params)
	if err != nil {
		return nil, err
	}

	for _, connectPeer := range resp.ConnectPeers {
		resources = append(resources, &NetworkManagerConnectPeer{
			svc:  svc,
			peer: connectPeer,
		})
	}

	return resources, nil
}

type NetworkManagerConnectPeer struct {
	svc  *networkmanager.NetworkManager
	peer *networkmanager.ConnectPeerSummary
}

func (n *NetworkManagerConnectPeer) Remove(_ context.Context) error {
	params := &networkmanager.DeleteConnectPeerInput{
		ConnectPeerId: n.peer.ConnectPeerId,
	}

	_, err := n.svc.DeleteConnectPeer(params)
	if err != nil {
		return err
	}

	return nil
}

func (n *NetworkManagerConnectPeer) Filter() error {
	if strings.EqualFold(ptr.ToString(n.peer.ConnectPeerState), awsutil.StateDeleted) {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (n *NetworkManagerConnectPeer) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", n.peer.ConnectPeerId)

	for _, tagValue := range n.peer.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (n *NetworkManagerConnectPeer) String() string {
	return *n.peer.ConnectPeerId
}
