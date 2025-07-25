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

func TestNetworkFirewallLoggingConfiguration_Remove(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockNetworkFirewall := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)

	resource := &NetworkFirewallLoggingConfiguration{
		svc:          mockNetworkFirewall,
		FirewallArn:  ptr.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
		FirewallName: ptr.String("test-firewall"),
		LoggingConfig: &networkfirewall.LoggingConfiguration{
			LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{
				{
					LogType:            aws.String("FLOW"),
					LogDestinationType: aws.String("S3"),
				},
			},
		},
	}

	mockNetworkFirewall.EXPECT().UpdateLoggingConfiguration(gomock.Any()).Return(&networkfirewall.UpdateLoggingConfigurationOutput{}, nil)

	err := resource.Remove(context.Background())
	a.Nil(err)
}

func TestNetworkFirewallLoggingConfigurationLister_List(t *testing.T) {
	a := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)
	lister := &NetworkFirewallLoggingConfigurationLister{
		mockSvc: mockSvc,
	}

	opts := &nuke.ListerOpts{
		AccountID: aws.String("123456789012"),
	}

	mockSvc.EXPECT().
		ListFirewallsPages(
			&networkfirewall.ListFirewallsInput{
				MaxResults: aws.Int64(100),
			},
			gomock.Any(),
		).
		DoAndReturn(func(input *networkfirewall.ListFirewallsInput,
			fn func(*networkfirewall.ListFirewallsOutput, bool) bool) error {
			fn(&networkfirewall.ListFirewallsOutput{
				Firewalls: []*networkfirewall.FirewallMetadata{
					{
						FirewallArn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
						FirewallName: aws.String("test-firewall"),
					},
				},
			}, true)
			return nil
		})

	mockSvc.EXPECT().
		DescribeLoggingConfiguration(&networkfirewall.DescribeLoggingConfigurationInput{
			FirewallArn: aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
		}).
		Return(&networkfirewall.DescribeLoggingConfigurationOutput{
			LoggingConfiguration: &networkfirewall.LoggingConfiguration{
				LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{
					{
						LogType:            aws.String("FLOW"),
						LogDestinationType: aws.String("S3"),
					},
				},
			},
		}, nil)

	resources, err := lister.List(context.Background(), opts)
	a.Nil(err)
	a.Len(resources, 1)

	loggingConfig := resources[0].(*NetworkFirewallLoggingConfiguration)
	a.Equal("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall", *loggingConfig.FirewallArn)
	a.Equal("test-firewall", *loggingConfig.FirewallName)
	a.Len(loggingConfig.LoggingConfig.LogDestinationConfigs, 1)
	a.Equal("FLOW", *loggingConfig.LoggingConfig.LogDestinationConfigs[0].LogType)
	a.Equal("S3", *loggingConfig.LoggingConfig.LogDestinationConfigs[0].LogDestinationType)
}
