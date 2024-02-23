# Usage

## aws-nuke

```console
NAME:
   aws-nuke - remove everything from an aws account

USAGE:
   aws-nuke [global options] command [command options] 

VERSION:
   3.0.0-dev

AUTHOR:
   Erik Kristensen <erik@erikkristensen.com>

COMMANDS:
   run, nuke                       run nuke against an aws account and remove everything from it
   account-details, account        list details about the AWS account that the tool is authenticated to
   config-details                  explain the configuration file and the resources that will be nuked
   resource-types, list-resources  list available resources to nuke
   help, h                         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## aws-nuke run

```console
NAME:
   aws-nuke run - run nuke against an aws account and remove everything from it

USAGE:
   aws-nuke run [command options] [arguments...]

OPTIONS:
   --config value                                                       path to config file (default: "config.yaml")
   --include value, --target value [ --include value, --target value ]  only run against these resource types
   --exclude value [ --exclude value ]                                  exclude these resource types
   --cloud-control value [ --cloud-control value ]                      use these resource types with the Cloud Control API instead of the default
   --quiet                                                              hide filtered messages (default: false)
   --no-dry-run                                                         actually run the removal of the resources after discovery (default: false)
   --no-alias-check                                                     disable aws account alias check - requires entry in config as well (default: false)
   --no-prompt, --force                                                 disable prompting for verification to run (default: false)
   --prompt-delay value, --force-sleep value                            seconds to delay after prompt before running (minimum: 3 seconds) (default: 10)
   --feature-flag value [ --feature-flag value ]                        enable experimental behaviors that may not be fully tested or supported
   --log-level value, -l value                                          Log Level (default: "info") [$LOGLEVEL]
   --log-caller                                                         log the caller (aka line number and file) (default: false)
   --log-disable-color                                                  disable log coloring (default: false)
   --log-full-timestamp                                                 force log output to always show full timestamp (default: false)
   --help, -h                                                           show help  
```

## aws-nuke explain-account

This command shows you details of how you are authenticated to AWS. 

```console
NAME:
   aws-nuke explain-account - explain the account and authentication method used to authenticate against AWS

USAGE:
   aws-nuke explain-account [command options] [arguments...]

DESCRIPTION:
   explain the account and authentication method used to authenticate against AWS

OPTIONS:
   --config value, -c value          path to config file (default: "config.yaml")
   --default-region value            the default aws region to use when setting up the aws auth session [$AWS_DEFAULT_REGION]
   --access-key-id value             the aws access key id to use when setting up the aws auth session [$AWS_ACCESS_KEY_ID]
   --secret-access-key value         the aws secret access key to use when setting up the aws auth session [$AWS_SECRET_ACCESS_KEY]
   --session-token value             the aws session token to use when setting up the aws auth session, typically used for temporary credentials [$AWS_SESSION_TOKEN]
   --profile value                   the aws profile to use when setting up the aws auth session, typically used for shared credentials files [$AWS_PROFILE]
   --assume-role-arn value           the role arn to assume using the credentials provided in the profile or statically set [$AWS_ASSUME_ROLE_ARN]
   --assume-role-session-name value  the session name to provide for the assumed role [$AWS_ASSUME_ROLE_SESSION_NAME]
   --assume-role-external-id value   the external id to provide for the assumed role [$AWS_ASSUME_ROLE_EXTERNAL_ID]
   --log-level value, -l value       Log Level (default: "info") [$LOGLEVEL]
   --log-caller                      log the caller (aka line number and file) (default: false)
   --log-disable-color               disable log coloring (default: false)
   --log-full-timestamp              force log output to always show full timestamp (default: false)
   --help, -h                        show help
```

### explain-account example output

```console
Overview:
> Account ID:       123456789012
> Account ARN:      arn:aws:iam::123456789012:root
> Account UserID:   AKIAIOSFODNN7EXAMPLE:root
> Account Alias:    no-alias-123456789012
> Default Region:   us-east-2
> Enabled Regions:  [global ap-south-1 ca-central-1 eu-central-1 us-west-1 us-west-2 eu-north-1 eu-west-3 eu-west-2 eu-west-1 ap-northeast-3 ap-northeast-2 ap-northeast-1 sa-east-1 ap-southeast-1 ap-southeast-2 us-east-1 us-east-2]

Authentication:
> Method: Static Keys
> Access Key ID:    AKIAIOSFODNN7EXAMPLE
```

## aws-nuke explain-config

This command will explain the configuration file and the resources that will be nuked for the targeted account. 

```console
NAME:
   aws-nuke explain-config - explain the configuration file and the resources that will be nuked for an account

USAGE:
   aws-nuke explain-config [command options] [arguments...]

DESCRIPTION:
   explain the configuration file and the resources that will be nuked for an account that
   is defined within the configuration. You may either specific an account using the --account-id flag or
   leave it empty to use the default account that can be authenticated against. If you want to see the
   resource types that will be nuked, use the --with-resource-types flag. If you want to see the resources
   that have filters defined, use the --with-resource-filters flag.

OPTIONS:
   --config value, -c value  path to config file (default: "config.yaml")
   --account-id value        the account id to check against the configuration file, if empty, it will use whatever account
      can be authenticated against
   --with-resource-filters           include resource with filters defined in the output (default: false)
   --with-resource-types             include resource types defined in the output (default: false)
   --default-region value            the default aws region to use when setting up the aws auth session [$AWS_DEFAULT_REGION]
   --access-key-id value             the aws access key id to use when setting up the aws auth session [$AWS_ACCESS_KEY_ID]
   --secret-access-key value         the aws secret access key to use when setting up the aws auth session [$AWS_SECRET_ACCESS_KEY]
   --session-token value             the aws session token to use when setting up the aws auth session, typically used for temporary credentials [$AWS_SESSION_TOKEN]
   --profile value                   the aws profile to use when setting up the aws auth session, typically used for shared credentials files [$AWS_PROFILE]
   --assume-role-arn value           the role arn to assume using the credentials provided in the profile or statically set [$AWS_ASSUME_ROLE_ARN]
   --assume-role-session-name value  the session name to provide for the assumed role [$AWS_ASSUME_ROLE_SESSION_NAME]
   --assume-role-external-id value   the external id to provide for the assumed role [$AWS_ASSUME_ROLE_EXTERNAL_ID]
   --log-level value, -l value       Log Level (default: "info") [$LOGLEVEL]
   --log-caller                      log the caller (aka line number and file) (default: false)
   --log-disable-color               disable log coloring (default: false)
   --log-full-timestamp              force log output to always show full timestamp (default: false)
   --help, -h                        show help
```

### explain-config example output

```console
Configuration Details

Resource Types:   426
Filter Presets:   2
Resource Filters: 24

Note: use --with-resource-filters to see resources with filters defined
Note: use --with-resource-types to see included resource types that will be nuked
```
