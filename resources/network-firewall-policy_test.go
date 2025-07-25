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

func TestNetworkFirewallPolicy_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)
	resource := &NetworkFirewallPolicy{
		svc:        mockSvc,
		PolicyArn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall-policy/test-policy"),
		PolicyName: aws.String("test-policy"),
	}

	mockSvc.EXPECT().
		DeleteFirewallPolicy(&networkfirewall.DeleteFirewallPolicyInput{
			FirewallPolicyArn: aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall-policy/test-policy"),
		}).
		Return(&networkfirewall.DeleteFirewallPolicyOutput{}, nil)

	err := resource.Remove(context.Background())
	a.Nil(err)
}

func TestNetworkFirewallPolicyLister_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)
	lister := &NetworkFirewallPolicyLister{
		mockSvc: mockSvc,
	}

	opts := &nuke.ListerOpts{
		AccountID: aws.String("123456789012"),
	}

	mockSvc.EXPECT().
		ListFirewallPoliciesPages(
			&networkfirewall.ListFirewallPoliciesInput{
				MaxResults: aws.Int64(100),
			},
			gomock.Any(),
		).
		DoAndReturn(func(input *networkfirewall.ListFirewallPoliciesInput,
			fn func(*networkfirewall.ListFirewallPoliciesOutput, bool) bool) error {
			fn(&networkfirewall.ListFirewallPoliciesOutput{
				FirewallPolicies: []*networkfirewall.FirewallPolicyMetadata{
					{
						Arn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall-policy/test-policy"),
						Name: aws.String("test-policy"),
					},
				},
			}, true)
			return nil
		})

	resources, err := lister.List(context.Background(), opts)
	a.Nil(err)
	a.Len(resources, 1)

	policy := resources[0].(*NetworkFirewallPolicy)
	a.Equal("arn:aws:network-firewall:us-west-2:123456789012:firewall-policy/test-policy", *policy.PolicyArn)
	a.Equal("test-policy", *policy.PolicyName)
}
