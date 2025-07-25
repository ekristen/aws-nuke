package resources

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_networkfirewalliface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func TestNetworkFirewallRuleGroup_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)
	resource := &NetworkFirewallRuleGroup{
		svc:           mockSvc,
		RuleGroupArn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:stateful-rulegroup/test-rule-group"),
		RuleGroupName: aws.String("test-rule-group"),
	}

	mockSvc.EXPECT().
		DeleteRuleGroup(&networkfirewall.DeleteRuleGroupInput{
			RuleGroupArn: aws.String("arn:aws:network-firewall:us-west-2:123456789012:stateful-rulegroup/test-rule-group"),
		}).
		Return(&networkfirewall.DeleteRuleGroupOutput{}, nil)

	err := resource.Remove(context.Background())
	a.Nil(err)
}

func TestNetworkFirewallRuleGroupLister_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)
	lister := &NetworkFirewallRuleGroupLister{
		mockSvc: mockSvc,
	}

	opts := &nuke.ListerOpts{
		AccountID: aws.String("123456789012"),
	}

	mockSvc.EXPECT().
		ListRuleGroupsPages(
			&networkfirewall.ListRuleGroupsInput{
				MaxResults: aws.Int64(100),
			},
			gomock.Any(),
		).
		DoAndReturn(func(input *networkfirewall.ListRuleGroupsInput,
			fn func(*networkfirewall.ListRuleGroupsOutput, bool) bool) error {
			fn(&networkfirewall.ListRuleGroupsOutput{
				RuleGroups: []*networkfirewall.RuleGroupMetadata{
					{
						Arn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:stateful-rulegroup/test-rule-group"),
						Name: aws.String("test-rule-group"),
					},
				},
			}, true)
			return nil
		})

	resources, err := lister.List(context.Background(), opts)
	a.Nil(err)
	a.Len(resources, 1)

	ruleGroup := resources[0].(*NetworkFirewallRuleGroup)
	a.Equal("arn:aws:network-firewall:us-west-2:123456789012:stateful-rulegroup/test-rule-group", *ruleGroup.RuleGroupArn)
	a.Equal("test-rule-group", *ruleGroup.RuleGroupName)
}
