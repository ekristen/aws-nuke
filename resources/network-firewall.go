package resources

import (
	"context"
	"fmt"

	"github.com/gotidy/ptr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/networkfirewall/networkfirewalliface"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const NetworkFirewallResource = "NetworkFirewall"

func init() {
	registry.Register(&registry.Registration{
		Name:                NetworkFirewallResource,
		Scope:               nuke.Account,
		Resource:            &NetworkFirewall{},
		Lister:              &NetworkFirewallLister{},
		AlternativeResource: "AWS::NetworkFirewall::Firewall",
	})
}

type NetworkFirewall struct {
	svc       networkfirewalliface.NetworkFirewallAPI
	accountID *string
	logger    *logrus.Entry

	ARN           *string `description:"The ARN of the firewall."`
	Name          *string `description:"The name of the firewall."`
	FirewallID    *string `description:"The ID of the firewall."`
	VPCID         *string `description:"The VPC ID where the firewall is deployed."`
	Status        *string `description:"The current status of the firewall."`
	Tags          []*networkfirewall.Tag
	LoggingConfig *networkfirewall.LoggingConfiguration
}

type NetworkFirewallLister struct {
	mockSvc networkfirewalliface.NetworkFirewallAPI
}

func (l *NetworkFirewallLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
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
				describeParams := &networkfirewall.DescribeFirewallInput{
					FirewallArn: firewall.FirewallArn,
				}

				describeOutput, err := svc.DescribeFirewall(describeParams)
				if err != nil {
					if opts.Logger != nil {
						opts.Logger.WithError(errors.WithStack(err)).
							WithField("firewall-arn", ptr.ToString(firewall.FirewallArn)).
							Error("failed to describe firewall")
					}
					continue
				}

				loggingParams := &networkfirewall.DescribeLoggingConfigurationInput{
					FirewallArn: firewall.FirewallArn,
				}
				loggingOutput, err := svc.DescribeLoggingConfiguration(loggingParams)
				var loggingConfig *networkfirewall.LoggingConfiguration
				if err != nil {
					if opts.Logger != nil {
						opts.Logger.WithError(errors.WithStack(err)).
							WithField("firewall-arn", ptr.ToString(firewall.FirewallArn)).
							Warn("failed to describe logging configuration, proceeding without it")
					}
				} else {
					loggingConfig = loggingOutput.LoggingConfiguration
				}

				var logger *logrus.Entry
				if opts.Logger != nil {
					logger = opts.Logger.WithField("firewall-name", ptr.ToString(firewall.FirewallName))
				}
				resources = append(resources, &NetworkFirewall{
					svc:           svc,
					accountID:     opts.AccountID,
					logger:        logger,
					ARN:           firewall.FirewallArn,
					Name:          firewall.FirewallName,
					FirewallID:    describeOutput.Firewall.FirewallId,
					VPCID:         describeOutput.Firewall.VpcId,
					Status:        describeOutput.FirewallStatus.Status,
					Tags:          describeOutput.Firewall.Tags,
					LoggingConfig: loggingConfig,
				})
			}
			return !lastPage
		}); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *NetworkFirewall) Filter() error {
	if ptr.ToString(r.Status) == "DELETING" {
		return fmt.Errorf("firewall is already being deleted")
	}
	return nil
}

func (r *NetworkFirewall) Remove(_ context.Context) error {
	if r.logger != nil {
		r.logger.Info("starting firewall deletion process")
	}

	if r.LoggingConfig != nil && len(r.LoggingConfig.LogDestinationConfigs) > 0 {
		if r.logger != nil {
			r.logger.Info("removing logging configuration before firewall deletion")
		}
		updateParams := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn: r.ARN,
			LoggingConfiguration: &networkfirewall.LoggingConfiguration{
				LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{},
			},
		}
		_, err := r.svc.UpdateLoggingConfiguration(updateParams)
		if err != nil {
			return errors.Wrap(err, "failed to remove logging configuration")
		}
		if r.logger != nil {
			r.logger.Info("logging configuration removed successfully")
		}
	}

	deleteParams := &networkfirewall.DeleteFirewallInput{
		FirewallArn: r.ARN,
	}
	_, err := r.svc.DeleteFirewall(deleteParams)
	if err != nil {
		return errors.Wrap(err, "failed to delete firewall")
	}

	if r.logger != nil {
		r.logger.Info("firewall deletion initiated successfully")
	}
	return nil
}

func (r *NetworkFirewall) Properties() types.Properties {
	props := types.NewPropertiesFromStruct(r)
	if r.LoggingConfig != nil && len(r.LoggingConfig.LogDestinationConfigs) > 0 {
		props.Set("HasLoggingConfiguration", true)
		props.Set("LogDestinationConfigsCount", len(r.LoggingConfig.LogDestinationConfigs))
	} else {
		props.Set("HasLoggingConfiguration", false)
	}
	return props
}

func (r *NetworkFirewall) String() string {
	return ptr.ToString(r.Name)
}
