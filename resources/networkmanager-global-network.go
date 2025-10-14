package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"                    //nolint:staticcheck
	"github.com/aws/aws-sdk-go/service/networkmanager" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NetworkManagerGlobalNetworkResource = "NetworkManagerGlobalNetwork"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkManagerGlobalNetworkResource,
		Scope:    nuke.Account,
		Resource: &NetworkManagerGlobalNetwork{},
		Lister:   &NetworkManagerGlobalNetworkLister{},
	})
}

type NetworkManagerGlobalNetworkLister struct{}

func (l *NetworkManagerGlobalNetworkLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := networkmanager.New(opts.Session)
	resources := make([]resource.Resource, 0)

	params := &networkmanager.DescribeGlobalNetworksInput{
		MaxResults: aws.Int64(100),
	}

	for {
		resp, err := svc.DescribeGlobalNetworks(params)
		if err != nil {
			return nil, err
		}

		for _, network := range resp.GlobalNetworks {
			resources = append(resources, &NetworkManagerGlobalNetwork{
				svc:     svc,
				network: network,
			})
		}
		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return resources, nil
}

type NetworkManagerGlobalNetwork struct {
	svc     *networkmanager.NetworkManager
	network *networkmanager.GlobalNetwork
}

func (n *NetworkManagerGlobalNetwork) Remove(_ context.Context) error {
	params := &networkmanager.DeleteGlobalNetworkInput{
		GlobalNetworkId: n.network.GlobalNetworkId,
	}

	_, err := n.svc.DeleteGlobalNetwork(params)
	if err != nil {
		return err
	}

	return nil
}

func (n *NetworkManagerGlobalNetwork) Filter() error {
	if strings.EqualFold(ptr.ToString(n.network.State), awsutil.StateDeleted) {
		return fmt.Errorf("already deleted")
	}

	return nil
}

func (n *NetworkManagerGlobalNetwork) Properties() types.Properties {
	properties := types.NewProperties()

	properties.Set("ID", n.network.GlobalNetworkId)
	properties.Set("ARN", n.network.GlobalNetworkArn)

	for _, tagValue := range n.network.Tags {
		properties.SetTag(tagValue.Key, tagValue.Value)
	}

	return properties
}

func (n *NetworkManagerGlobalNetwork) String() string {
	return *n.network.GlobalNetworkId
}
