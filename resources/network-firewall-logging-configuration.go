package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/networkfirewall/networkfirewalliface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NetworkFirewallLoggingConfigurationResource = "NetworkFirewallLoggingConfiguration"

func init() {
	registry.Register(&registry.Registration{
		Name:                NetworkFirewallLoggingConfigurationResource,
		Scope:               nuke.Account,
		Resource:            &NetworkFirewallLoggingConfiguration{},
		Lister:              &NetworkFirewallLoggingConfigurationLister{},
		AlternativeResource: "AWS::NetworkFirewall::LoggingConfiguration",
	})
}

type NetworkFirewallLoggingConfigurationLister struct {
	mockSvc networkfirewalliface.NetworkFirewallAPI
}

func (l *NetworkFirewallLoggingConfigurationLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	resources := make([]resource.Resource, 0)

	var svc networkfirewalliface.NetworkFirewallAPI
	if l.mockSvc != nil {
		svc = l.mockSvc
	} else {
		svc = networkfirewall.New(opts.Session)
	}

	params := &networkfirewall.ListFirewallsInput{
		MaxResults: aws.Int64(100),
	}

	if err := svc.ListFirewallsPages(params,
		func(page *networkfirewall.ListFirewallsOutput, lastPage bool) bool {
			for _, firewall := range page.Firewalls {
				loggingParams := &networkfirewall.DescribeLoggingConfigurationInput{
					FirewallArn: firewall.FirewallArn,
				}
				loggingOutput, err := svc.DescribeLoggingConfiguration(loggingParams)
				if err != nil {
					if opts.Logger != nil {
						opts.Logger.WithError(err).
							WithField("firewall-arn", ptr.ToString(firewall.FirewallArn)).
							Warn("failed to describe logging configuration, skipping")
					}
					continue
				}

				if loggingOutput.LoggingConfiguration != nil && len(loggingOutput.LoggingConfiguration.LogDestinationConfigs) > 0 {
					resources = append(resources, &NetworkFirewallLoggingConfiguration{
						svc:           svc,
						accountID:     opts.AccountID,
						FirewallArn:   firewall.FirewallArn,
						FirewallName:  firewall.FirewallName,
						LoggingConfig: loggingOutput.LoggingConfiguration,
					})
				}
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

type NetworkFirewallLoggingConfiguration struct {
	svc           networkfirewalliface.NetworkFirewallAPI
	accountID     *string
	FirewallArn   *string `description:"The ARN of the firewall."`
	FirewallName  *string `description:"The name of the firewall."`
	LoggingConfig *networkfirewall.LoggingConfiguration
}

func (r *NetworkFirewallLoggingConfiguration) Filter() error {
	return nil
}

func (r *NetworkFirewallLoggingConfiguration) Remove(_ context.Context) error {
	updateParams := &networkfirewall.UpdateLoggingConfigurationInput{
		FirewallArn: r.FirewallArn,
		LoggingConfiguration: &networkfirewall.LoggingConfiguration{
			LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{},
		},
	}
	_, err := r.svc.UpdateLoggingConfiguration(updateParams)
	return err
}

func (r *NetworkFirewallLoggingConfiguration) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	if r.LoggingConfig != nil && len(r.LoggingConfig.LogDestinationConfigs) > 0 {
		props.Set("HasLoggingConfiguration", true)
		props.Set("LogDestinationConfigsCount", len(r.LoggingConfig.LogDestinationConfigs))
	}
	return props
}

func (r *NetworkFirewallLoggingConfiguration) String() string {
	return fmt.Sprintf("%s (Firewall: %s)", NetworkFirewallLoggingConfigurationResource, ptr.ToString(r.FirewallName))
}
