package nuke

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	libconfig "github.com/ekristen/libnuke/pkg/config"
	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/scanner"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/common"
	"github.com/ekristen/aws-nuke/v3/pkg/config"
	"github.com/ekristen/aws-nuke/v3/pkg/nuke"

	"github.com/ekristen/aws-nuke/v3/resources"
)

// ConfigureCreds is a helper function to configure the awsutil.Credentials object from the cli.Context
func ConfigureCreds(c *cli.Command) (creds *awsutil.Credentials) {
	creds = &awsutil.Credentials{}

	creds.Profile = c.String("profile")
	creds.AccessKeyID = c.String("access-key-id")
	creds.SecretAccessKey = c.String("secret-access-key")
	creds.SessionToken = c.String("session-token")
	creds.AssumeRoleArn = c.String("assume-role-arn")
	creds.RoleSessionName = c.String("assume-role-session-name")
	creds.ExternalID = c.String("assume-role-external-id")

	return creds
}

func execute(baseCtx context.Context, c *cli.Command) error { //nolint:funlen,gocyclo
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	defaultRegion := c.String("default-region")
	creds := ConfigureCreds(c)

	if err := creds.Validate(); err != nil {
		return err
	}

	// Create the parameters object that will be used to configure the nuke process.
	params := &libnuke.Parameters{
		Force:             c.Bool("force"),
		ForceSleep:        c.Int("force-sleep"),
		Quiet:             c.Bool("quiet"),
		NoDryRun:          c.Bool("no-dry-run"),
		Includes:          c.StringSlice("include"),
		Excludes:          c.StringSlice("exclude"),
		Alternatives:      c.StringSlice("cloud-control"),
		MaxWaitRetries:    c.Int("max-wait-retries"),
		MaxFailureRetries: c.Int("max-failure-retries"),
	}

	if len(c.StringSlice("feature-flag")) > 0 {
		if slices.Contains(c.StringSlice("feature-flag"), "wait-on-dependencies") {
			params.WaitOnDependencies = true
		}

		if slices.Contains(c.StringSlice("feature-flag"), "filter-groups") {
			params.UseFilterGroups = true
		}
	}

	logger := logrus.StandardLogger()
	logger.SetOutput(os.Stdout)

	// Parse the user supplied configuration file to pass in part to configure the nuke process.
	parsedConfig, err := config.New(libconfig.Options{
		Path:         c.String("config"),
		Deprecations: registry.GetDeprecatedResourceTypeMapping(),
		Log:          logger.WithField("component", "config"),
	})
	if err != nil {
		logger.Errorf("Failed to parse config file %s", c.String("config"))
		return err
	}

	// Set the default region for the AWS SDK to use.
	if defaultRegion != "" {
		awsutil.DefaultRegionID = defaultRegion

		partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), defaultRegion)
		if !ok {
			if parsedConfig.CustomEndpoints.GetRegion(defaultRegion) == nil {
				err = fmt.Errorf(
					"the custom region '%s' must be specified in the configuration 'endpoints'"+
						" to determine its partition", defaultRegion)
				logger.WithError(err).Errorf("unable to resolve partition for region: %s", defaultRegion)
				return err
			}
		}

		awsutil.DefaultAWSPartitionID = partition.ID()
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

	n.SetRunSleep(c.Duration("run-sleep-delay"))
	n.SetLogger(logger.WithField("component", "libnuke"))
	n.RegisterVersion(common.AppVersion.String())

	// Register our custom validate handler that validates the account and AWS nuke unique alias checks
	n.RegisterValidateHandler(func() error {
		return parsedConfig.ValidateAccount(account.ID(), account.Aliases(), c.Bool("no-alias-check"))
	})

	// Register our custom prompt handler that shows the account information
	p := &nuke.Prompt{Parameters: params, Account: account, Logger: logger}
	n.RegisterPrompt(p.Prompt)

	// Get any specific account level configuration
	accountConfig := parsedConfig.Accounts[account.ID()]

	// Get current registered resource names
	resourceNames := registry.GetNames()

	// Combine all the places where alternative resource types can be defined and then dynamically
	// register them as a Cloud Control resource type.
	altResourceTypes := types.Collection(registry.ExpandNames(n.Parameters.Alternatives))
	altResourceTypes = altResourceTypes.Union(parsedConfig.ResourceTypes.GetAlternatives())
	altResourceTypes = altResourceTypes.Union(accountConfig.ResourceTypes.GetAlternatives())
	for _, rt := range altResourceTypes {
		if slices.Contains(resourceNames, rt) {
			continue
		}

		resources.RegisterCloudControl(rt)
	}

	// Resolve the resource types to be used for the nuke process based on the parameters, global configuration, and
	// account level configuration.
	resourceTypes := types.ResolveResourceTypes(
		registry.GetNames(), // note: we want to re-pull the registry here due to the dynamic registration above
		[]types.Collection{
			registry.ExpandNames(n.Parameters.Includes),
			parsedConfig.ResourceTypes.GetIncludes(),
			accountConfig.ResourceTypes.GetIncludes(),
		},
		[]types.Collection{
			registry.ExpandNames(n.Parameters.Excludes),
			parsedConfig.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			registry.ExpandNames(n.Parameters.Alternatives),
			parsedConfig.ResourceTypes.GetAlternatives(),
			accountConfig.ResourceTypes.GetAlternatives(),
		},
		registry.GetAlternativeResourceTypeMapping(),
	)

	// If the user has specified the "all" region, then we need to get the enabled regions for the account
	// and use those. Otherwise, we will use the regions that are specified in the configuration.
	if slices.Contains(parsedConfig.Regions, "all") {
		parsedConfig.Regions = account.Regions()

		logger.Info(
			`"all" detected in region list, only enabled regions and "global" will be used, all others ignored`)

		if len(parsedConfig.Regions) > 1 {
			logger.Warnf(`additional regions defined along with "all", these will be ignored!`)
		}

		logger.Infof("The following regions are enabled for the account (%d total):", len(parsedConfig.Regions))

		printableRegions := make([]string, 0)
		for i, region := range parsedConfig.Regions {
			printableRegions = append(printableRegions, region)
			if i%6 == 0 { // print 5 regions per line
				logger.Infof("> %s", strings.Join(printableRegions, ", "))
				printableRegions = make([]string, 0)
			} else if i == len(parsedConfig.Regions)-1 {
				logger.Infof("> %s", strings.Join(printableRegions, ", "))
			}
		}
	}

	// Register the scanners for each region that is defined in the configuration.
	for _, regionName := range parsedConfig.Regions {
		// Step 1 - Create the region object
		region := nuke.NewRegion(regionName, account.ResourceTypeToServiceType, account.NewSession, account.NewConfig)

		// Step 2 - Create the scannerActual object
		scannerActual, scannerActualErr := scanner.New(&scanner.Config{
			Owner:         regionName,
			ResourceTypes: resourceTypes,
			Opts: &nuke.ListerOpts{
				Region:    region,
				AccountID: ptr.String(account.ID()),
				Logger: logger.WithFields(logrus.Fields{
					"component": "scanner",
					"region":    regionName,
				}),
			},
			Logger:          logger,
			ParallelQueries: c.Int64("parallel-queries"),
			QueueSize:       c.Int("max-queue-size"),
		})
		if scannerActualErr != nil {
			return scannerActualErr
		}

		// Step 3 - Register a mutate function that will be called to modify the lister options for each resource type
		// see pkg/nuke/resource.go for the MutateOpts function. Its purpose is to create the proper session for the
		// proper region.
		regMutateErr := scannerActual.RegisterMutateOptsFunc(nuke.MutateOpts)
		if regMutateErr != nil {
			return regMutateErr
		}

		// Step 4 - Register the scannerActual with the nuke object
		regScanErr := n.RegisterScanner(nuke.Account, scannerActual)
		if regScanErr != nil {
			return regScanErr
		}
	}

	return n.Run(ctx)
}

func init() { //nolint:funlen
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "path to config file",
			Value:   "config.yaml",
			Action:  common.CheckFilePath,
		},
		&cli.StringSliceFlag{
			Name:    "include",
			Usage:   "only run against these resource types",
			Aliases: []string{"target"},
		},
		&cli.StringSliceFlag{
			Name:    "exclude",
			Aliases: []string{"exclude-resource"},
			Usage:   "exclude these resource types",
		},
		&cli.StringSliceFlag{
			Name:  "cloud-control",
			Usage: "use these resource types with the Cloud Control API instead of the default",
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "hide filtered messages",
		},
		&cli.BoolFlag{
			Name:  "no-dry-run",
			Usage: "actually run the removal of the resources after discovery",
		},
		&cli.BoolFlag{
			Name:    "no-prompt",
			Usage:   "disable prompting for verification to run",
			Aliases: []string{"force"},
		},
		&cli.IntFlag{
			Name:    "prompt-delay",
			Usage:   "seconds to delay after prompt before running (minimum: 3 seconds)",
			Value:   10,
			Aliases: []string{"force-sleep"},
			Action:  common.CheckRealInt,
		},
		&cli.IntFlag{
			Name:   "max-wait-retries",
			Usage:  "maximum number of retries to wait for dependencies to be removed",
			Action: common.CheckRealInt,
		},
		&cli.IntFlag{
			Name:   "max-failure-retries",
			Usage:  "maximum number of retries to wait for failed dependencies to be removed",
			Action: common.CheckRealInt,
		},
		&cli.DurationFlag{
			Name:    "run-sleep-delay",
			Sources: cli.EnvVars("AWS_NUKE_RUN_SLEEP_DELAY"),
			Usage:   "time to sleep between run/loops of resource deletions, default is 5 seconds",
			Value:   5 * time.Second,
		},
		&cli.BoolFlag{
			Name:  "no-alias-check",
			Usage: "disable aws account alias check - requires entry in config as well",
		},
		&cli.StringSliceFlag{
			Name:  "feature-flag",
			Usage: "enable experimental behaviors that may not be fully tested or supported",
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
		&cli.IntFlag{
			Name:    "parallel-queries",
			Usage:   "CAUTION! ADVANCED USAGE! number of parallel resource queries to run at a time",
			Sources: cli.EnvVars("AWS_NUKE_PARALLEL_QUERIES"),
			Value:   scanner.DefaultParallelQueries,
			Hidden:  true,
		},
		&cli.IntFlag{
			Name:    "max-queue-size",
			Usage:   "CAUTION! ADVANCED USAGE! the max number of items to queue up before aws-nuke will error",
			Sources: cli.EnvVars("AWS_NUKE_MAX_QUEUE_SIZE"),
			Value:   scanner.DefaultQueueSize,
			Hidden:  true,
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
