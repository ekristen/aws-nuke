package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/networkfirewall/networkfirewalliface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NetworkFirewallResource = "NetworkFirewall"

func init() {
	registry.Register(&registry.Registration{
		Name:     NetworkFirewallResource,
		Scope:    nuke.Account,
		Resource: &NetworkFirewall{},
		Lister:   &NetworkFirewallLister{},
		DependsOn: []string{
			NetworkFirewallLoggingConfigurationResource,
		},
		AlternativeResource: "AWS::NetworkFirewall::Firewall",
	})
}

type NetworkFirewallLister struct {
	mockSvc networkfirewalliface.NetworkFirewallAPI
}

func (l *NetworkFirewallLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc networkfirewalliface.NetworkFirewallAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = networkfirewall.New(opts.Session)
	}

	params := &networkfirewall.ListFirewallsInput{
		MaxResults: aws.Int64(100),
	}

	if err := svc.ListFirewallsPages(params,
		func(page *networkfirewall.ListFirewallsOutput, lastPage bool) bool {
			for _, firewall := range page.Firewalls {
				resources = append(resources, &NetworkFirewall{
					svc:          svc,
					accountID:    opts.AccountID,
					FirewallArn:  firewall.FirewallArn,
					FirewallName: firewall.FirewallName,
				})
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

type NetworkFirewall struct {
	svc          networkfirewalliface.NetworkFirewallAPI
	accountID    *string
	FirewallArn  *string `description:"The ARN of the firewall."`
	FirewallName *string `description:"The name of the firewall."`
}

func (r *NetworkFirewall) Remove(_ context.Context) error {
	_, err := r.svc.DeleteFirewall(&networkfirewall.DeleteFirewallInput{
		FirewallArn: r.FirewallArn,
	})
	return err
}

func (r *NetworkFirewall) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkFirewall) String() string {
	return ptr.ToString(r.FirewallName)
}
