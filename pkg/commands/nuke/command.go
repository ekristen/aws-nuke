package nuke

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	libconfig "github.com/ekristen/libnuke/pkg/config"
	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/pkg/awsutil"
	"github.com/ekristen/aws-nuke/pkg/commands/global"
	"github.com/ekristen/aws-nuke/pkg/common"
	"github.com/ekristen/aws-nuke/pkg/config"
	"github.com/ekristen/aws-nuke/pkg/nuke"
)

func execute(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	var (
		err           error
		creds         awsutil.Credentials
		defaultRegion string
	)

	if !creds.HasKeys() && !creds.HasProfile() && defaultRegion != "" {
		creds.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		creds.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	err = creds.Validate()
	if err != nil {
		return err
	}

	// Create the parameters object that will be used to configure the nuke process.
	params := &libnuke.Parameters{
		Force:        c.Bool("force"),
		ForceSleep:   c.Int("force-sleep"),
		Quiet:        c.Bool("quiet"),
		NoDryRun:     c.Bool("no-dry-run"),
		Includes:     c.StringSlice("only-resource"),
		Excludes:     c.StringSlice("exclude-resource"),
		Alternatives: c.StringSlice("cloud-control"),
	}

	// Parse the user supplied configuration file to pass in part to configure the nuke process.
	parsedConfig, err := config.New(libconfig.Options{
		Path:         c.Path("config"),
		Deprecations: resource.GetDeprecatedResourceTypeMapping(),
	})
	if err != nil {
		logrus.Errorf("Failed to parse config file %s", c.Path("config"))
		return err
	}

	// Set the default region for the AWS SDK to use.
	if defaultRegion != "" {
		awsutil.DefaultRegionID = defaultRegion
		switch defaultRegion {
		case endpoints.UsEast1RegionID, endpoints.UsEast2RegionID, endpoints.UsWest1RegionID, endpoints.UsWest2RegionID:
			awsutil.DefaultAWSPartitionID = endpoints.AwsPartitionID
		case endpoints.UsGovEast1RegionID, endpoints.UsGovWest1RegionID:
			awsutil.DefaultAWSPartitionID = endpoints.AwsUsGovPartitionID
		default:
			if parsedConfig.CustomEndpoints.GetRegion(defaultRegion) == nil {
				err = fmt.Errorf("the custom region '%s' must be specified in the configuration 'endpoints'", defaultRegion)
				logrus.Error(err.Error())
				return err
			}
		}
	}

	// Create the AWS Account object. This will be used to get the account ID and aliases for the account.
	account, err := awsutil.NewAccount(creds, parsedConfig.CustomEndpoints)
	if err != nil {
		return err
	}

	// Get the filters for the account that is being connected to via the AWS SDK.
	filters, err := parsedConfig.Filters(account.ID())
	if err != nil {
		return err
	}

	// Instantiate libnuke
	n := libnuke.New(params, filters, parsedConfig.Settings)

	// Register our custom validate handler that validates the account and AWS nuke unique alias checks
	n.RegisterValidateHandler(func() error {
		return parsedConfig.ValidateAccount(account.ID(), account.Aliases())
	})

	// Register our custom prompt handler that shows the account information
	p := &nuke.Prompt{Parameters: params, Account: account}
	n.RegisterPrompt(p.Prompt)

	// Get any specific account level configuration
	accountConfig := parsedConfig.Accounts[account.ID()]

	// Resolve the resource types to be used for the nuke process based on the parameters, global configuration, and
	// account level configuration.
	resourceTypes := types.ResolveResourceTypes(
		resource.GetNames(),
		[]types.Collection{
			n.Parameters.Includes,
			parsedConfig.ResourceTypes.Targets,
			accountConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			parsedConfig.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			n.Parameters.Alternatives,
			parsedConfig.ResourceTypes.CloudControl,
			accountConfig.ResourceTypes.CloudControl,
		},
		resource.GetAlternativeResourceTypeMapping(),
	)

	// Register the scanners for each region that is defined in the configuration.
	for _, regionName := range parsedConfig.Regions {
		// Step 1 - Create the region object
		region := nuke.NewRegion(regionName, account.ResourceTypeToServiceType, account.NewSession)

		// Step 2 - Create the scanner object
		scanner := libnuke.NewScanner(regionName, resourceTypes, &nuke.ListerOpts{
			Region: region,
		})

		// Step 3 - Register a mutate function that will be called to modify the lister options for each resource type
		// see pkg/nuke/resource.go for the MutateOpts function. Its purpose is to create the proper session for the
		// proper region.
		regMutateErr := scanner.RegisterMutateOptsFunc(nuke.MutateOpts)
		if regMutateErr != nil {
			return regMutateErr
		}

		// Step 4 - Register the scanner with the nuke object
		regScanErr := n.RegisterScanner(nuke.Account, scanner)
		if regScanErr != nil {
			return regScanErr
		}
	}

	return n.Run(ctx)
}

func init() {
	flags := []cli.Flag{
		&cli.PathFlag{
			Name:  "config",
			Usage: "path to config file",
			Value: "config.yaml",
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "disable prompting for verification to run",
		},
		&cli.IntFlag{
			Name:  "force-sleep",
			Usage: "seconds to sleep",
			Value: 10,
		},
		&cli.BoolFlag{
			Name:  "quiet",
			Usage: "hide filtered messages",
		},
		&cli.BoolFlag{
			Name:  "no-dry-run",
			Usage: "actually run the removal of the resources after discovery",
		},
		&cli.StringSliceFlag{
			Name:    "only-resource",
			Usage:   "only run against these resource types",
			Aliases: []string{"target", "include", "include-resource"},
		},
		&cli.StringSliceFlag{
			Name:    "exclude-resource",
			Usage:   "exclude these resource types",
			Aliases: []string{"exclude"},
		},
		&cli.StringSliceFlag{
			Name:  "cloud-control",
			Usage: "use these resource types with the Cloud Control API instead of the default",
		},
	}

	cmd := &cli.Command{
		Name:  "run",
		Usage: "run nuke against an aws account and remove everything from it",
		Aliases: []string{
			"nuke",
		},
		Flags:  append(flags, global.Flags()...),
		Before: global.Before,
		Action: execute,
	}

	common.RegisterCommand(cmd)
}
