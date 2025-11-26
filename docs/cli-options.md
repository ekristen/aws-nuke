# Options

This is not a comprehensive list of options, but rather a list of features that I think are worth highlighting.

## Cloud Control API

`--cloud-control` will allow you to use the Cloud Control API for specific resource types. This is useful if you want to use the Cloud Control API for specific resource types.

## Skip Alias Checks

`--no-alias-check` will skip the check for the AWS account alias. This is useful if you are running in an account that does not have an alias.

## Skip Prompts

`--no-prompt` will skip the prompt to verify you want to run the command. This is useful if you are running in a CI/CD environment.
`--prompt-delay` will set the delay before the command runs. This is useful if you want to give yourself time to cancel the command.

## Logging

- `--log-level` will set the log level. This is useful if you want to see more or less information in the logs.
- `--log-caller` will log the caller (aka line number and file). This is useful if you are debugging.
- `--log-disable-color` will disable log coloring. This is useful if you are running in an environment that does not support color.
- `--log-full-timestamp` will force log output to always show full timestamp. This is useful if you want to see the full timestamp in the logs.
- `--log-format=json` (or `--json`) will output logs as JSON. This is useful for parsing output and generating reports.
