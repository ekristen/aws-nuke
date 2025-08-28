package resources

import (
	"context"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"

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

type NetworkFirewallLister struct{}

func (l *NetworkFirewallLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := networkfirewall.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &networkfirewall.ListFirewallsInput{
		MaxResults: aws.Int32(100),
	}

	paginator := networkfirewall.NewListFirewallsPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, firewall := range page.Firewalls {
			resources = append(resources, &NetworkFirewall{
				svc:       svc,
				accountID: opts.AccountID,
				ARN:       firewall.FirewallArn,
				Name:      firewall.FirewallName,
			})
		}
	}

	return resources, nil
}

type NetworkFirewall struct {
	svc       *networkfirewall.Client
	accountID *string
	ARN       *string `description:"The ARN of the firewall."`
	Name      *string `description:"The name of the firewall."`
}

func (r *NetworkFirewall) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteFirewall(ctx, &networkfirewall.DeleteFirewallInput{
		FirewallArn: r.ARN,
	})
	return err
}

func (r *NetworkFirewall) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkFirewall) String() string {
	return ptr.ToString(r.Name)
}
