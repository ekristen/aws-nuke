package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	networkfirewalltypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"

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

type NetworkFirewallLoggingConfigurationLister struct{}

func (l *NetworkFirewallLoggingConfigurationLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)
	svc := networkfirewall.NewFromConfig(*opts.Config)
	resources := make([]resource.Resource, 0)

	params := &networkfirewall.ListFirewallsInput{
		MaxResults: aws.Int32(100),
	}

	paginator := networkfirewall.NewListFirewallsPaginator(svc, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, firewall := range page.Firewalls {
			loggingParams := &networkfirewall.DescribeLoggingConfigurationInput{
				FirewallArn: firewall.FirewallArn,
			}
			loggingOutput, err := svc.DescribeLoggingConfiguration(ctx, loggingParams)
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
	}

	return resources, nil
}

type NetworkFirewallLoggingConfiguration struct {
	svc           *networkfirewall.Client
	accountID     *string
	FirewallArn   *string `description:"The ARN of the firewall."`
	FirewallName  *string `description:"The name of the firewall."`
	LoggingConfig *networkfirewalltypes.LoggingConfiguration
}

func (r *NetworkFirewallLoggingConfiguration) Filter() error {
	return nil
}

func (r *NetworkFirewallLoggingConfiguration) Remove(ctx context.Context) error {
	updateParams := &networkfirewall.UpdateLoggingConfigurationInput{
		FirewallArn: r.FirewallArn,
		LoggingConfiguration: &networkfirewalltypes.LoggingConfiguration{
			LogDestinationConfigs: []networkfirewalltypes.LogDestinationConfig{},
		},
	}
	_, err := r.svc.UpdateLoggingConfiguration(ctx, updateParams)
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
	return fmt.Sprintf("%s -> %s", ptr.ToString(r.FirewallName), NetworkFirewallLoggingConfigurationResource)
}
