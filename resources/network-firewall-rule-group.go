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

const NetworkFirewallRuleGroupResource = "NetworkFirewallRuleGroup"

func init() {
	registry.Register(&registry.Registration{
		Name:                NetworkFirewallRuleGroupResource,
		Scope:               nuke.Account,
		Resource:            &NetworkFirewallRuleGroup{},
		Lister:              &NetworkFirewallRuleGroupLister{},
		AlternativeResource: "AWS::NetworkFirewall::RuleGroup",
	})
}

type NetworkFirewallRuleGroupLister struct {
	mockSvc networkfirewalliface.NetworkFirewallAPI
}

func (l *NetworkFirewallRuleGroupLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc networkfirewalliface.NetworkFirewallAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = networkfirewall.New(opts.Session)
	}

	params := &networkfirewall.ListRuleGroupsInput{
		MaxResults: aws.Int64(100),
	}

	if err := svc.ListRuleGroupsPages(params,
		func(page *networkfirewall.ListRuleGroupsOutput, lastPage bool) bool {
			for _, ruleGroup := range page.RuleGroups {
				resources = append(resources, &NetworkFirewallRuleGroup{
					svc:           svc,
					accountID:     opts.AccountID,
					RuleGroupArn:  ruleGroup.Arn,
					RuleGroupName: ruleGroup.Name,
				})
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

type NetworkFirewallRuleGroup struct {
	svc           networkfirewalliface.NetworkFirewallAPI
	accountID     *string
	RuleGroupArn  *string `description:"The ARN of the rule group."`
	RuleGroupName *string `description:"The name of the rule group."`
}

func (r *NetworkFirewallRuleGroup) Remove(_ context.Context) error {
	_, err := r.svc.DeleteRuleGroup(&networkfirewall.DeleteRuleGroupInput{
		RuleGroupArn: r.RuleGroupArn,
	})
	return err
}

func (r *NetworkFirewallRuleGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkFirewallRuleGroup) String() string {
	return ptr.ToString(r.RuleGroupName)
}
