# Usage

## aws-nuke

```console
NAME:
   aws-nuke - remove everything from an aws account

USAGE:
   aws-nuke [global options] command [command options] 

VERSION:
   3.0.0-beta.2

AUTHOR:
   Erik Kristensen <erik@erikkristensen.com>

COMMANDS:
   run, nuke                       run nuke against an aws account and remove everything from it
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