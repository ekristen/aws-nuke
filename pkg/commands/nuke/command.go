package nuke

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	sdknuke "github.com/ekristen/libnuke/pkg/nuke"
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

	_ = ctx

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

	params := nuke.Parameters{
		Parameters: sdknuke.Parameters{
			Force:      c.Bool("force"),
			ForceSleep: c.Int("force-sleep"),
			Quiet:      c.Bool("quiet"),
			NoDryRun:   c.Bool("no-dry-run"),
		},
		Targets:      c.StringSlice("only-resource"),
		Excludes:     c.StringSlice("exclude-resource"),
		CloudControl: c.StringSlice("cloud-control"),
	}

	parsedConfig, err := config.Load(c.Path("config"))
	if err != nil {
		logrus.Errorf("Failed to parse config file %s", c.Path("config"))
		return err
	}

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

	account, err := awsutil.NewAccount(creds, parsedConfig.CustomEndpoints)
	if err != nil {
		return err
	}

	filters, err := parsedConfig.Filters(account.ID())
	if err != nil {
		return err
	}

	n := nuke.New(params, parsedConfig, filters, *account)

	n.RegisterValidateHandler(func() error {
		return parsedConfig.ValidateAccount(n.Account.ID(), n.Account.Aliases())
	})

	n.RegisterPrompt(n.Prompt)

	accountConfig := parsedConfig.Accounts[n.Account.ID()]
	resourceTypes := nuke.ResolveResourceTypes(
		resource.GetNames(),
		nuke.GetCloudControlMapping(),
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.GetResourceTypes().Targets,
			accountConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.GetResourceTypes().Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{
			n.Parameters.CloudControl,
			n.Config.GetResourceTypes().CloudControl,
			accountConfig.ResourceTypes.CloudControl,
		},
	)

	// mutateOps is a function that will be called for each resource type to mutate the options
	// for the scanner based on whatever criteria you want. However, in this case for the aws-nuke
	// tool, it's mutating the opts to create the proper session for the proper region.
	var mutateOps = func(opts interface{}, resourceType string) interface{} {
		o := opts.(*nuke.ListerOpts)

		session, err := o.Region.Session(resourceType)
		if err != nil {
			panic(err)
		}

		o.Session = session
		return o
	}

	for _, regionName := range parsedConfig.Regions {
		region := nuke.NewRegion(regionName, n.Account.ResourceTypeToServiceType, n.Account.NewSession)
		scanner := sdknuke.NewScanner(regionName, resourceTypes, &nuke.ListerOpts{
			Region: region,
		})

		regMutateErr := scanner.RegisterMutateOptsFunc(mutateOps)
		if regMutateErr != nil {
			return regMutateErr
		}

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
			Aliases: []string{"target"},
		},
		&cli.BoolFlag{
			Name:  "experimental-deps",
			Usage: "turn on dependency removal ordering",
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
