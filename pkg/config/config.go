package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/settings"
)

// New creates a new extended configuration from a file. This is necessary because we are extended the default
// libnuke configuration to contain additional attributes that are specific to the AWS Nuke tool.
func New(opts config.Options) (*Config, error) {
	// Step 1 - Create the libnuke config
	cfg, err := config.New(opts)
	if err != nil {
		return nil, err
	}

	// Step 2 - Instantiate the extended config
	c := &Config{
		CustomEndpoints: make(CustomEndpoints, 0),
	}

	// Step 3 - Load the same config file against the extended config
	// Intentionally ignored, this will never error because we already validated the file exists
	_ = c.Load(opts.Path)

	// Step 4 - Set the libnuke config on the extended config
	c.Config = cfg

	// Step 5 - Resolve any deprecated feature flags
	c.ResolveDeprecatedFeatureFlags()

	return c, nil
}

// Config is an extended configuration implementation that adds some additional features on top of the libnuke config.
type Config struct {
	// Config is the underlying libnuke configuration.
	*config.Config `yaml:",inline"`

	// BypassAliasCheckAccounts is a list of account IDs that will be allowed to bypass the alias check.
	// This is useful for accounts that don't have an alias for a number of reasons, it must be used with a cli
	// flag --no-alias-check to be effective.
	BypassAliasCheckAccounts []string `yaml:"bypass-alias-check-accounts"`

	// FeatureFlags is a collection of feature flags that can be used to enable or disable certain behaviors on
	// resources. This is left over from the AWS Nuke tool and is deprecated. It was left to make the transition to the
	// library and ekristen/aws-nuke@v3 easier for existing users.
	// Deprecated: Use Settings instead. Will be removed in 4.x
	FeatureFlags *FeatureFlags `yaml:"feature-flags"`

	// CustomEndpoints is a collection of custom endpoints that can be used to override the default AWS endpoints.
	CustomEndpoints CustomEndpoints `yaml:"endpoints"`
}

// Load loads a configuration from a file and parses it into a Config struct.
func (c *Config) Load(path string) error {
	var err error

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(raw, c); err != nil {
		return err
	}

	return nil
}

// InBypassAliasCheckAccounts returns true if the specified account ID is in the bypass alias check accounts list.
func (c *Config) InBypassAliasCheckAccounts(accountID string) bool {
	for _, id := range c.BypassAliasCheckAccounts {
		if id == accountID {
			return true
		}
	}

	return false
}

// ValidateAccount validates the account ID and aliases for the specified account. This will return an error if the
// account ID is invalid, the account ID is blocklisted, the account doesn't have an alias, the account alias contains
// the substring 'prod', or the account ID isn't listed in the config.
func (c *Config) ValidateAccount(accountID string, aliases []string, skipAliasChecks bool) error {
	// Call the libnuke config validation first
	if err := c.Config.ValidateAccount(accountID); err != nil {
		return err
	}

	if skipAliasChecks {
		if c.InBypassAliasCheckAccounts(accountID) {
			return nil
		}

		c.Log.Warnf("--no-alias-check is set, but the account ID '%s' isn't in the bypass list.", accountID)
	}

	if len(aliases) == 0 {
		return fmt.Errorf("specified account doesn't have an alias. " +
			"For safety reasons you need to specify an account alias. " +
			"Your production account should contain the term 'prod'")
	}

	for _, alias := range aliases {
		if strings.Contains(strings.ToLower(alias), "prod") {
			return fmt.Errorf("you are trying to nuke an account with the alias '%s', "+
				"but it has the substring 'prod' in it. Aborting", alias)
		}
	}

	return nil
}

// ResolveDeprecatedFeatureFlags resolves any deprecated feature flags in the configuration. This converts the legacy
// feature flags into the new settings format. The feature flags will be deprecated with version 4.x. This was left in
// place to make the transition to the libnuke library and ekristen/aws-nuke@v3 easier for existing users.
func (c *Config) ResolveDeprecatedFeatureFlags() {
	if c.FeatureFlags != nil {
		c.Log.Warn("deprecated configuration key 'feature-flags' - please use 'settings' instead")

		if c.FeatureFlags.ForceDeleteLightsailAddOns {
			c.Settings.Set("LightsailInstance", &settings.Setting{
				"ForceDeleteAddOns": true,
			})
		}
		if c.FeatureFlags.DisableEC2InstanceStopProtection {
			c.Settings.Set("EC2Instance", &settings.Setting{
				"DisableStopProtection": true,
			})
		}
		if c.FeatureFlags.DisableDeletionProtection.EC2Instance {
			c.Settings.Set("EC2Instance", &settings.Setting{
				"DisableDeletionProtection": true,
			})
		}
		if c.FeatureFlags.DisableDeletionProtection.RDSInstance {
			c.Settings.Set("RDSInstance", &settings.Setting{
				"DisableDeletionProtection": true,
			})
		}
		if c.FeatureFlags.DisableDeletionProtection.ELBv2 {
			c.Settings.Set("ELBv2", &settings.Setting{
				"DisableDeletionProtection": true,
			})
		}
		if c.FeatureFlags.DisableDeletionProtection.CloudformationStack {
			c.Settings.Set("CloudFormationStack", &settings.Setting{
				"DisableDeletionProtection": true,
			})
		}
		if c.FeatureFlags.DisableDeletionProtection.QLDBLedger {
			c.Settings.Set("QLDBLedger", &settings.Setting{
				"DisableDeletionProtection": true,
			})
		}
	}
}

// FeatureFlags is a collection of feature flags that can be used to enable or disable certain features of the nuke
// This is left over from the AWS Nuke tool and is deprecated. It was left to make the transition to the library and
// ekristen/aws-nuke@v3 easier for existing users.
// Deprecated: Use Settings instead. Will be removed in 4.x
type FeatureFlags struct {
	DisableDeletionProtection        DisableDeletionProtection `yaml:"disable-deletion-protection"`
	DisableEC2InstanceStopProtection bool                      `yaml:"disable-ec2-instance-stop-protection"`
	ForceDeleteLightsailAddOns       bool                      `yaml:"force-delete-lightsail-addons"`
}

// DisableDeletionProtection is a collection of feature flags that can be used to disable deletion protection for
// certain resource types. This is left over from the AWS Nuke tool and is deprecated. It was left to make transition
// to the library and ekristen/aws-nuke@v3 easier for existing users.
// Deprecated: Use Settings instead. Will be removed in 4.x
type DisableDeletionProtection struct {
	RDSInstance         bool `yaml:"RDSInstance"`
	EC2Instance         bool `yaml:"EC2Instance"`
	CloudformationStack bool `yaml:"CloudformationStack"`
	ELBv2               bool `yaml:"ELBv2"`
	QLDBLedger          bool `yaml:"QLDBLedger"`
}

// CustomService is a custom service endpoint that can be used to override the default AWS endpoints.
type CustomService struct {
	Service               string `yaml:"service"`
	URL                   string `yaml:"url"`
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify"`
}

// CustomServices is a collection of custom service endpoints that can be used to override the default AWS endpoints.
type CustomServices []*CustomService

// CustomRegion is a custom region endpoint that can be used to override the default AWS regions
type CustomRegion struct {
	Region                string         `yaml:"region"`
	Services              CustomServices `yaml:"services"`
	TLSInsecureSkipVerify bool           `yaml:"tls_insecure_skip_verify"`
}

// CustomEndpoints is a collection of custom region endpoints that can be used to override the default AWS regions
type CustomEndpoints []*CustomRegion

// GetRegion returns the custom region or nil when no such custom endpoints are defined for this region
func (endpoints CustomEndpoints) GetRegion(region string) *CustomRegion {
	for _, r := range endpoints {
		if r.Region == region {
			if r.TLSInsecureSkipVerify {
				for _, s := range r.Services {
					s.TLSInsecureSkipVerify = r.TLSInsecureSkipVerify
				}
			}
			return r
		}
	}
	return nil
}

// GetService returns the custom region or nil when no such custom endpoints are defined for this region
func (services CustomServices) GetService(serviceType string) *CustomService {
	for _, s := range services {
		if serviceType == s.Service {
			return s
		}
	}
	return nil
}

// GetURL returns the custom region or nil when no such custom endpoints are defined for this region
func (endpoints CustomEndpoints) GetURL(region, serviceType string) string {
	r := endpoints.GetRegion(region)
	if r == nil {
		return ""
	}
	s := r.Services.GetService(serviceType)
	if s == nil {
		return ""
	}
	return s.URL
}
