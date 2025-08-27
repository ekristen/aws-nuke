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

type NetworkFirewallRuleGroupLister struct{}

func (l *NetworkFirewallRuleGroupLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := networkfirewall.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &networkfirewall.ListRuleGroupsInput{
		MaxResults: aws.Int32(100),
	}

	paginator := networkfirewall.NewListRuleGroupsPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, ruleGroup := range page.RuleGroups {
			resources = append(resources, &NetworkFirewallRuleGroup{
				svc:       svc,
				accountID: opts.AccountID,
				ARN:       ruleGroup.Arn,
				Name:      ruleGroup.Name,
			})
		}
	}

	return resources, nil
}

type NetworkFirewallRuleGroup struct {
	svc       *networkfirewall.Client
	accountID *string
	ARN       *string `description:"The ARN of the rule group."`
	Name      *string `description:"The name of the rule group."`
}

func (r *NetworkFirewallRuleGroup) Remove(ctx context.Context) error {
	_, err := r.svc.DeleteRuleGroup(ctx, &networkfirewall.DeleteRuleGroupInput{
		RuleGroupArn: r.ARN,
	})
	return err
}

func (r *NetworkFirewallRuleGroup) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *NetworkFirewallRuleGroup) String() string {
	return ptr.ToString(r.Name)
}
