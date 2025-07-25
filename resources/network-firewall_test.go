package resources

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"

	"github.com/ekristen/aws-nuke/v3/mocks/mock_networkfirewalliface"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

func TestNetworkFirewall_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNetworkFirewall := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)

	resource := &NetworkFirewall{
		svc:          mockNetworkFirewall,
		FirewallArn:  ptr.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
		FirewallName: ptr.String("test-firewall"),
	}

	mockNetworkFirewall.EXPECT().DeleteFirewall(gomock.Any()).Return(&networkfirewall.DeleteFirewallOutput{}, nil)

	err := resource.Remove(context.Background())
	a.Nil(err)
}

func TestNetworkFirewallLister_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNetworkFirewall := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)

	lister := &NetworkFirewallLister{
		mockSvc: mockNetworkFirewall,
	}

	opts := &nuke.ListerOpts{
		AccountID: ptr.String("123456789012"),
	}

	mockNetworkFirewall.EXPECT().ListFirewallsPages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(input *networkfirewall.ListFirewallsInput, fn func(*networkfirewall.ListFirewallsOutput, bool) bool) error {
			output := &networkfirewall.ListFirewallsOutput{
				Firewalls: []*networkfirewall.FirewallMetadata{
					{
						FirewallArn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
						FirewallName: aws.String("test-firewall"),
					},
				},
			}
			fn(output, true)
			return nil
		})

	resources, err := lister.List(context.Background(), opts)
	a.Nil(err)
	a.Len(resources, 1)

	firewall := resources[0].(*NetworkFirewall)
	a.Equal("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall", ptr.ToString(firewall.FirewallArn))
	a.Equal("test-firewall", ptr.ToString(firewall.FirewallName))
}
