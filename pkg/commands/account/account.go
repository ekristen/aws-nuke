package account

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"github.com/aws/aws-sdk-go/aws/endpoints" //nolint:staticcheck

	libconfig "github.com/ekristen/libnuke/pkg/config"
	"github.com/ekristen/libnuke/pkg/registry"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/commands/nuke"
	"github.com/ekristen/aws-nuke/v3/pkg/common"
	"github.com/ekristen/aws-nuke/v3/pkg/config"
)

func execute(_ context.Context, c *cli.Command) error {
	defaultRegion := c.String("default-region")
	creds := nuke.ConfigureCreds(c)

	if err := creds.Validate(); err != nil {
		return err
	}

	// Parse the user supplied configuration file to pass in part to configure the nuke process.
	parsedConfig, err := config.New(libconfig.Options{
		Path:         c.String("config"),
		Deprecations: registry.GetDeprecatedResourceTypeMapping(),
	})
	if err != nil {
		logrus.Errorf("Failed to parse config file %s", c.String("config"))
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
				logrus.WithError(err).Errorf("unable to resolve partition for region: %s", defaultRegion)
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

	fmt.Println("Overview:")
	fmt.Println("> Account ID:      ", account.ID())
	fmt.Println("> Account ARN:     ", account.ARN())
	fmt.Println("> Account UserID:  ", account.UserID())
	fmt.Println("> Account Alias:   ", account.Alias())
	fmt.Println("> Default Region:  ", defaultRegion)
	fmt.Println("> Enabled Regions: ", account.Regions())

	fmt.Println("")
	fmt.Println("Authentication:")
	if creds.HasKeys() {
		fmt.Println("> Method: Static Keys")
		fmt.Println("> Access Key ID:   ", creds.AccessKeyID)
	}
	if creds.HasProfile() {
		fmt.Println("> Method: Shared Credentials")
		fmt.Println("> Profile:         ", creds.Profile)
	}
	if creds.AssumeRoleArn != "" {
		fmt.Println("> Method: Assume Role")
		fmt.Println("> Role ARN:        ", creds.AssumeRoleArn)
		if creds.RoleSessionName != "" {
			fmt.Println("> Session Name:    ", creds.RoleSessionName)
		}
		if creds.ExternalID != "" {
			fmt.Println("> External ID:     ", creds.ExternalID)
		}
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
		Name:        "explain-account",
		Usage:       "explain the account and authentication method used to authenticate against AWS",
		Description: `explain the account and authentication method used to authenticate against AWS`,
		Flags:       append(flags, global.Flags()...),
		Before:      global.Before,
		Action:      execute,
	}

	common.RegisterCommand(cmd)
}
