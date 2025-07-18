package config

import (
	"context"
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	libconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/commands/nuke"
	"github.com/ekristen/aws-nuke/v3/pkg/common"
	"github.com/ekristen/aws-nuke/v3/pkg/config"
)

func execute(_ context.Context, c *cli.Command) error { //nolint:funlen,gocyclo
	accountID := c.String("account-id")

	parsedConfig, err := config.New(libconfig.Options{
		Path:         c.String("config"),
		Deprecations: registry.GetDeprecatedResourceTypeMapping(),
	})
	if err != nil {
		logrus.Errorf("Failed to parse config file %s", c.String("config"))
		return err
	}

	if accountID == "" {
		logrus.Info("no account id provided, attempting to authenticate and get account id")
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
		return fmt.Errorf("account %s is not configured in the config file", accountID)
	}

	// Resolve the resource types to be used for the nuke process based on the parameters, global configuration, and
	// account level configuration.
	resourceTypes := types.ResolveResourceTypes(
		registry.GetNames(),
		[]types.Collection{
			{}, // note: empty collection since we are not capturing parameters
			parsedConfig.ResourceTypes.GetIncludes(),
			accountConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			{}, // note: empty collection since we are not capturing parameters
			parsedConfig.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			{}, // note: empty collection since we are not capturing parameters
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

	fmt.Printf("Account ID:       %s\n", accountID)
	fmt.Printf("Resource Types:   %d (total)\n", len(registry.GetNames()))
	fmt.Printf("      Included:   %d\n", len(resourceTypes))
	fmt.Printf("      Excluded:   %d\n", len(registry.GetNames())-len(resourceTypes))
	fmt.Printf("Filter Presets:   %d\n", len(accountConfig.Presets))
	fmt.Printf("Resource Filters: %d\n", filtersTotal)

	fmt.Println("")

	if c.Bool("with-filtered") {
		fmt.Println("Resources with Filters Defined:")
		for _, resource := range resourcesWithFilters {
			fmt.Printf("  %s\n", resource)
		}
		fmt.Println("")
	}

	if c.Bool("with-included") {
		fmt.Println("Resource Types:")
		for _, resourceType := range resourceTypes {
			fmt.Printf("  %s\n", resourceType)
		}
		fmt.Println("")
	}

	if c.Bool("with-excluded") {
		fmt.Println("Excluded Resource Types:")
		for _, resourceType := range registry.GetNames() {
			if !slices.Contains(resourceTypes, resourceType) {
				fmt.Printf("  %s\n", resourceType)
			}
		}
		fmt.Println("")
	}

	if !c.Bool("with-filtered") {
		fmt.Printf("Note: use --with-filtered to see resources with filters defined\n")
	}
	if !c.Bool("with-included") {
		fmt.Printf("Note: use --with-included to see included resource types that will be nuked\n")
	}
	if !c.Bool("with-excluded") {
		fmt.Printf("Note: use --with-excluded to see excluded resource types\n")
	}

	return nil
}

func init() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "path to config file",
			Value:   "config.yaml",
			Action:  common.CheckFilePath,
		},
		&cli.StringFlag{
			Name:  "account-id",
			Usage: `the account id to check against the configuration file, if empty, it will use whatever account can be authenticated against`,
		},
		&cli.BoolFlag{
			Name:  "with-filtered",
			Usage: "print out resource types that have filters defined against them",
		},
		&cli.BoolFlag{
			Name:  "with-included",
			Usage: "print out the included resource types",
		},
		&cli.BoolFlag{
			Name:  "with-excluded",
			Usage: "print out the excluded resource types",
		},
		&cli.StringFlag{
			Name:    "default-region",
			Sources: cli.EnvVars("AWS_DEFAULT_REGION"),
			Usage:   "the default aws region to use when setting up the aws auth session",
		},
		&cli.StringFlag{
			Name:    "access-key-id",
			Sources: cli.EnvVars("AWS_ACCESS_KEY_ID"),
			Usage:   "the aws access key id to use when setting up the aws auth session",
		},
		&cli.StringFlag{
			Name:    "secret-access-key",
			Sources: cli.EnvVars("AWS_SECRET_ACCESS_KEY"),
			Usage:   "the aws secret access key to use when setting up the aws auth session",
		},
		&cli.StringFlag{
			Name:    "session-token",
			Sources: cli.EnvVars("AWS_SESSION_TOKEN"),
			Usage:   "the aws session token to use when setting up the aws auth session, typically used for temporary credentials",
		},
		&cli.StringFlag{
			Name:    "profile",
			Sources: cli.EnvVars("AWS_PROFILE"),
			Usage:   "the aws profile to use when setting up the aws auth session, typically used for shared credentials files",
		},
		&cli.StringFlag{
			Name:    "assume-role-arn",
			Sources: cli.EnvVars("AWS_ASSUME_ROLE_ARN"),
			Usage:   "the role arn to assume using the credentials provided in the profile or statically set",
		},
		&cli.StringFlag{
			Name:    "assume-role-session-name",
			Sources: cli.EnvVars("AWS_ASSUME_ROLE_SESSION_NAME"),
			Usage:   "the session name to provide for the assumed role",
		},
		&cli.StringFlag{
			Name:    "assume-role-external-id",
			Sources: cli.EnvVars("AWS_ASSUME_ROLE_EXTERNAL_ID"),
			Usage:   "the external id to provide for the assumed role",
		},
	}

	cmd := &cli.Command{
		Name:  "explain-config",
		Usage: "explain the configuration file and the resources that will be nuked for an account",
		Description: `explain the configuration file and the resources that will be nuked for an account that
is defined within the configuration. You may either specific an account using the --account-id flag or
leave it empty to use the default account that can be authenticated against. You can optionally list out included,
excluded and resources with filters with their respective with flags.`,
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
		Action: execute,
	}

	common.RegisterCommand(cmd)
}
