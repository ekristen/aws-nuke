package resources

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/nuke"
)

const ConfigServiceConfigRuleResource = "ConfigServiceConfigRule"

func init() {
	registry.Register(&registry.Registration{
		Name:   ConfigServiceConfigRuleResource,
		Scope:  nuke.Account,
		Lister: &ConfigServiceConfigRuleLister{},
	})
}

type ConfigServiceConfigRuleLister struct{}

func (l *ConfigServiceConfigRuleLister) List(_ context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	svc := configservice.New(opts.Session)
	var resources []resource.Resource

	params := &configservice.DescribeConfigRulesInput{}

	for {
		output, err := svc.DescribeConfigRules(params)
		if err != nil {
			return nil, err
		}

		for _, configRule := range output.ConfigRules {
			remConfig, err := svc.DescribeRemediationConfigurations(&configservice.DescribeRemediationConfigurationsInput{
				ConfigRuleNames: []*string{configRule.ConfigRuleName},
			})
			if err != nil {
				logrus.
					WithField("name", configRule.ConfigRuleName).
					WithError(err).
					Warn("unable to describe remediation configurations")
			}

			newResource := &ConfigServiceConfigRule{
				svc:       svc,
				Name:      configRule.ConfigRuleName,
				CreatedBy: configRule.CreatedBy,
			}

			if remConfig != nil && len(remConfig.RemediationConfigurations) > 0 {
				newResource.HasRemediationConfig = ptr.Bool(true)
			}

			resources = append(resources, newResource)
		}

		if output.NextToken == nil {
			break
		}

		params.NextToken = output.NextToken
	}

	return resources, nil
}

type ConfigServiceConfigRule struct {
	svc                  *configservice.ConfigService
	Name                 *string
	Scope                *string
	HasRemediationConfig *bool
	CreatedBy            *string
}

func (f *ConfigServiceConfigRule) Filter() error {
	if aws.StringValue(f.CreatedBy) == "securityhub.amazonaws.com" {
		return fmt.Errorf("cannot remove rule owned by securityhub.amazonaws.com")
	}

	if aws.StringValue(f.CreatedBy) == "config-conforms.amazonaws.com" {
		return fmt.Errorf("cannot remove rule owned by config-conforms.amazonaws.com")
	}

	return nil
}

func (f *ConfigServiceConfigRule) Remove(_ context.Context) error {
	if ptr.ToBool(f.HasRemediationConfig) {
		if _, err := f.svc.DeleteRemediationConfiguration(&configservice.DeleteRemediationConfigurationInput{
			ConfigRuleName: f.Name,
		}); err != nil {
			return err
		}
	}

	_, err := f.svc.DeleteConfigRule(&configservice.DeleteConfigRuleInput{
		ConfigRuleName: f.Name,
	})

	return err
}

func (f *ConfigServiceConfigRule) String() string {
	return *f.Name
}

func (f *ConfigServiceConfigRule) Properties() types.Properties {
	return types.NewPropertiesFromStruct(f)
}
