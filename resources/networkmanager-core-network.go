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

const NetworkManagerCoreNetworkResource = "NetworkManagerCoreNetwork"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkManagerCoreNetworkResource,
		Scope:    nuke.Account,
		Resource: &NetworkManagerCoreNetwork{},
		Lister:   &NetworkManagerCoreNetworkLister{},
	})
}

type NetworkManagerCoreNetworkLister struct{}

func (l *NetworkManagerCoreNetworkLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := networkmanager.New(opts.Session)
	params := &networkmanager.ListCoreNetworksInput{}
	resources := make([]resource.Resource, 0)

	resp, err := svc.ListCoreNetworks(params)
	if err != nil {
		return nil, err
	}

	for _, network := range resp.CoreNetworks {
		resources = append(resources, &NetworkManagerCoreNetwork{
			svc:     svc,
			network: network,
		})
	}

	return resources, nil
}

type NetworkManagerCoreNetwork struct {
	svc     *networkmanager.NetworkManager
	network *networkmanager.CoreNetworkSummary
}

func (n *NetworkManagerCoreNetwork) Remove(_ context.Context) error {
	params := &networkmanager.DeleteCoreNetworkInput{
		CoreNetworkId: n.network.CoreNetworkId,
	}

	_, err := n.svc.DeleteCoreNetwork(params)
	if err != nil {
		return err
	}

	return nil
}

func (n *NetworkManagerCoreNetwork) Filter() error {
	if strings.EqualFold(ptr.ToString(n.network.State), awsutil.StateDeleted) {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (n *NetworkManagerCoreNetwork) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", n.network.CoreNetworkId)
	properties.Set("ARN", n.network.CoreNetworkArn)

	for _, tagValue := range n.network.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (n *NetworkManagerCoreNetwork) String() string {
	return *n.network.CoreNetworkId
}
