package resources

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/ekristen/aws-nuke/v3/mocks/mock_networkfirewalliface"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ekristen/libnuke/pkg/registry"
)

func TestNetworkFirewall_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)

	// Mock ListFirewalls
	mockSvc.EXPECT().ListFirewallsPages(gomock.Any(), gomock.Any()).DoAndReturn(
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

	// Mock DescribeFirewall
	mockSvc.EXPECT().DescribeFirewall(gomock.Any()).Return(
		&networkfirewall.DescribeFirewallOutput{
			Firewall: &networkfirewall.Firewall{
				FirewallArn:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
				FirewallName: aws.String("test-firewall"),
				FirewallId:   aws.String("firewall-12345678"),
				VpcId:        aws.String("vpc-12345678"),
				Tags:         []*networkfirewall.Tag{},
			},
			FirewallStatus: &networkfirewall.FirewallStatus{
				Status: aws.String("READY"),
			},
		}, nil)

	// Mock DescribeLoggingConfiguration
	mockSvc.EXPECT().DescribeLoggingConfiguration(gomock.Any()).Return(
		&networkfirewall.DescribeLoggingConfigurationOutput{
			LoggingConfiguration: &networkfirewall.LoggingConfiguration{
				LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{},
			},
		}, nil)

	lister := &NetworkFirewallLister{
		mockSvc: mockSvc,
	}

	resources, err := lister.List(context.Background(), testListerOpts)
	assert.NoError(t, err)
	assert.Len(t, resources, 1)

	firewall := resources[0].(*NetworkFirewall)
	assert.Equal(t, "test-firewall", *firewall.Name)
	assert.Equal(t, "firewall-12345678", *firewall.FirewallID)
	assert.Equal(t, "vpc-12345678", *firewall.VPCID)
}

func TestNetworkFirewall_Remove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)

	firewall := &NetworkFirewall{
		svc:  mockSvc,
		ARN:  aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
		Name: aws.String("test-firewall"),
		LoggingConfig: &networkfirewall.LoggingConfiguration{
			LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{
				{
					LogDestination: map[string]*string{
						"bucketName": aws.String("test-bucket"),
					},
					LogDestinationType: aws.String("S3"),
					LogType:            aws.String("ALERT"),
				},
			},
		},
	}

	// Expect logging configuration update
	mockSvc.EXPECT().UpdateLoggingConfiguration(gomock.Any()).Return(
		&networkfirewall.UpdateLoggingConfigurationOutput{}, nil)

	// Expect firewall deletion
	mockSvc.EXPECT().DeleteFirewall(gomock.Any()).Return(
		&networkfirewall.DeleteFirewallOutput{}, nil)

	err := firewall.Remove(context.Background())
	assert.NoError(t, err)
}

func TestNetworkFirewall_Remove_NoLoggingConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_networkfirewalliface.NewMockNetworkFirewallAPI(ctrl)

	firewall := &NetworkFirewall{
		svc:           mockSvc,
		ARN:           aws.String("arn:aws:network-firewall:us-west-2:123456789012:firewall/test-firewall"),
		Name:          aws.String("test-firewall"),
		LoggingConfig: nil,
	}

	// Should not call UpdateLoggingConfiguration when no logging config exists
	// Expect only firewall deletion
	mockSvc.EXPECT().DeleteFirewall(gomock.Any()).Return(
		&networkfirewall.DeleteFirewallOutput{}, nil)

	err := firewall.Remove(context.Background())
	assert.NoError(t, err)
}

func TestNetworkFirewall_Properties(t *testing.T) {
	firewall := &NetworkFirewall{
		Name: aws.String("test-firewall"),
		LoggingConfig: &networkfirewall.LoggingConfiguration{
			LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{
				{
					LogDestination: map[string]*string{
						"bucketName": aws.String("test-bucket"),
					},
					LogDestinationType: aws.String("S3"),
					LogType:            aws.String("ALERT"),
				},
			},
		},
	}

	props := firewall.Properties()
	assert.Equal(t, "true", props.Get("HasLoggingConfiguration"))
	assert.Equal(t, "1", props.Get("LogDestinationConfigsCount"))
}

func TestNetworkFirewall_String(t *testing.T) {
	firewall := &NetworkFirewall{
		Name: aws.String("test-firewall"),
	}

	assert.Equal(t, "test-firewall", firewall.String())
}

func TestNetworkFirewall_Registration(t *testing.T) {
	reg := registry.GetRegistration(NetworkFirewallResource)
	assert.NotNil(t, reg)
	assert.Equal(t, NetworkFirewallResource, reg.Name)
	assert.Equal(t, "AWS::NetworkFirewall::Firewall", reg.AlternativeResource)
}
