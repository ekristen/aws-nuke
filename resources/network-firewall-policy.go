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

const NetworkFirewallPolicyResource = "NetworkFirewallPolicy"

func init() {
	registry.Register(&registry.Registration{
		Name:                NetworkFirewallPolicyResource,
		Scope:               nuke.Account,
		Resource:            &NetworkFirewallPolicy{},
		Lister:              &NetworkFirewallPolicyLister{},
		AlternativeResource: "AWS::NetworkFirewall::FirewallPolicy",
	})
}

type NetworkFirewallPolicyLister struct{}

func (l *NetworkFirewallPolicyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := networkfirewall.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &networkfirewall.ListFirewallPoliciesInput{
		MaxResults: aws.Int32(100),
	}

	paginator := networkfirewall.NewListFirewallPoliciesPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, policy := range page.FirewallPolicies {
			resources = append(resources, &NetworkFirewallPolicy{
				svc:       svc,
				accountID: opts.AccountID,
				ARN:       policy.Arn,
				Name:      policy.Name,
			})
		}
	}

	return resources, nil
}

type NetworkFirewallPolicy struct {
	svc       *networkfirewall.Client
	accountID *string
	ARN       *string `description:"The ARN of the firewall policy."`
	Name      *string `description:"The name of the firewall policy."`
}

func (r *NetworkFirewallPolicy) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteFirewallPolicy(ctx, &networkfirewall.DeleteFirewallPolicyInput{
		FirewallPolicyArn: r.ARN,
	})
	return err
}

func (r *NetworkFirewallPolicy) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkFirewallPolicy) String() string {
	return ptr.ToString(r.Name)
}
