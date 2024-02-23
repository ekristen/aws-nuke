package config

import (
	"fmt"
	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/commands/global"
	"github.com/ekristen/aws-nuke/pkg/commands/nuke"
	"github.com/ekristen/aws-nuke/pkg/common"
	"github.com/ekristen/aws-nuke/pkg/config"
	libconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"slices"
)

func execute(c *cli.Context) error {
	accountID := c.String("account-id")

	parsedConfig, err := config.New(libconfig.Options{
		Path:         c.Path("config"),
		Deprecations: registry.GetDeprecatedResourceTypeMapping(),
	})
	if err != nil {
		logrus.Errorf("Failed to parse config file %s", c.Path("config"))
		return err
	}

	if accountID != "" {
		creds := nuke.ConfigureCreds(c)
		if err := creds.Validate(); err != nil {
			return err
		}

		// Create the AWS Account object. This will be used to get the account ID and aliases for the account.
		account, err := awsutil.NewAccount(creds, parsedConfig.CustomEndpoints)
		if err != nil {
			return err
		}

		accountID = account.ID()
	}

	// Get any specific account level configuration
	accountConfig := parsedConfig.Accounts[accountID]

	if accountConfig == nil {
		return fmt.Errorf("account is not configured in the config file")
	}

	// Resolve the resource types to be used for the nuke process based on the parameters, global configuration, and
	// account level configuration.
	resourceTypes := types.ResolveResourceTypes(
		registry.GetNames(),
		[]types.Collection{
			types.Collection{}, // note: empty collection since we are not capturing parameters
			parsedConfig.ResourceTypes.GetIncludes(),
			accountConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			types.Collection{}, // note: empty collection since we are not capturing parameters
			parsedConfig.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			types.Collection{}, // note: empty collection since we are not capturing parameters
			parsedConfig.ResourceTypes.GetAlternatives(),
			accountConfig.ResourceTypes.GetAlternatives(),
		},
		registry.GetAlternativeResourceTypeMapping(),
	)

	filtersTotal := 0
	var resourcesWithFilters []string
	for name, preset := range parsedConfig.Presets {
		if !slices.Contains(accountConfig.Presets, name) {
			continue
		}

		filtersTotal += len(preset.Filters)

		for resource := range preset.Filters {
			resourcesWithFilters = append(resourcesWithFilters, resource)
		}
	}

	fmt.Printf("Configuration Details\n\n")

	fmt.Printf("Resource Types:   %d\n", len(resourceTypes))
	fmt.Printf("Filter Presets:   %d\n", len(accountConfig.Presets))
	fmt.Printf("Resource Filters: %d\n", filtersTotal)

	fmt.Println("")

	if c.Bool("with-resource-filters") {
		fmt.Println("Resources with Filters Defined:")
		for _, resource := range resourcesWithFilters {
			fmt.Printf("  %s\n", resource)
		}
		fmt.Println("")
	} else {
		fmt.Printf("Note: use --with-resource-filters to see resources with filters defined\n")
	}

	if c.Bool("with-resource-types") {
		fmt.Println("Resource Types:")
		for _, resourceType := range resourceTypes {
			fmt.Printf("  %s\n", resourceType)
		}
		fmt.Println("")
	} else {
		fmt.Printf("Note: use --with-resource-types to see included resource types that will be nuked\n")
	}

	return nil
}

func init() {
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "path to config file",
			Value:   "config.yaml",
		},
		&cli.StringFlag{
			Name:  "account-id",
			Usage: `the account id to check against the configuration file, if empty, it will use whatever account can be authenticated against`,
		},
		&cli.BoolFlag{
			Name:  "with-resource-filters",
			Usage: "include resource with filters defined in the output",
		},
		&cli.BoolFlag{
			Name:  "with-resource-types",
			Usage: "include resource types defined in the output",
		},
		&cli.StringFlag{
			Name:    "default-region",
			EnvVars: []string{"AWS_DEFAULT_REGION"},
			Usage:   "the default aws region to use when setting up the aws auth session",
		},
		&cli.StringFlag{
			Name:    "access-key-id",
			EnvVars: []string{"AWS_ACCESS_KEY_ID"},
			Usage:   "the aws access key id to use when setting up the aws auth session",
		},
		&cli.StringFlag{
			Name:    "secret-access-key",
			EnvVars: []string{"AWS_SECRET_ACCESS_KEY"},
			Usage:   "the aws secret access key to use when setting up the aws auth session",
		},
		&cli.StringFlag{
			Name:    "session-token",
			EnvVars: []string{"AWS_SESSION_TOKEN"},
			Usage:   "the aws session token to use when setting up the aws auth session, typically used for temporary credentials",
		},
		&cli.StringFlag{
			Name:    "profile",
			EnvVars: []string{"AWS_PROFILE"},
			Usage:   "the aws profile to use when setting up the aws auth session, typically used for shared credentials files",
		},
		&cli.StringFlag{
			Name:    "assume-role-arn",
			EnvVars: []string{"AWS_ASSUME_ROLE_ARN"},
			Usage:   "the role arn to assume using the credentials provided in the profile or statically set",
		},
		&cli.StringFlag{
			Name:    "assume-role-session-name",
			EnvVars: []string{"AWS_ASSUME_ROLE_SESSION_NAME"},
			Usage:   "the session name to provide for the assumed role",
		},
		&cli.StringFlag{
			Name:    "assume-role-external-id",
			EnvVars: []string{"AWS_ASSUME_ROLE_EXTERNAL_ID"},
			Usage:   "the external id to provide for the assumed role",
		},
	}

	cmd := &cli.Command{
		Name:  "explain-config",
		Usage: "explain the configuration file and the resources that will be nuked for an account",
		Description: `explain the configuration file and the resources that will be nuked for an account that
is defined within the configuration. You may either specific an account using the --account-id flag or
leave it empty to use the default account that can be authenticated against. If you want to see the
resource types that will be nuked, use the --with-resource-types flag. If you want to see the resources
that have filters defined, use the --with-resource-filters flag.`,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
		Action: execute,
	}

	common.RegisterCommand(cmd)
}
