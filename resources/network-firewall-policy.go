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

type NetworkFirewallPolicyLister struct {
	mockSvc networkfirewalliface.NetworkFirewallAPI
}

func (l *NetworkFirewallPolicyLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc networkfirewalliface.NetworkFirewallAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = networkfirewall.New(opts.Session)
	}

	params := &networkfirewall.ListFirewallPoliciesInput{
		MaxResults: aws.Int64(100),
	}

	if err := svc.ListFirewallPoliciesPages(params,
		func(page *networkfirewall.ListFirewallPoliciesOutput, lastPage bool) bool {
			for _, policy := range page.FirewallPolicies {
				resources = append(resources, &NetworkFirewallPolicy{
					svc:        svc,
					accountID:  opts.AccountID,
					PolicyArn:  policy.Arn,
					PolicyName: policy.Name,
				})
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

type NetworkFirewallPolicy struct {
	svc        networkfirewalliface.NetworkFirewallAPI
	accountID  *string
	PolicyArn  *string `description:"The ARN of the firewall policy."`
	PolicyName *string `description:"The name of the firewall policy."`
}

func (r *NetworkFirewallPolicy) Remove(_ context.Context) error {
	_, err := r.svc.DeleteFirewallPolicy(&networkfirewall.DeleteFirewallPolicyInput{
		FirewallPolicyArn: r.PolicyArn,
	})
	return err
}

func (r *NetworkFirewallPolicy) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkFirewallPolicy) String() string {
	return ptr.ToString(r.PolicyName)
}
