# Authentication

There are multiple ways to authenticate to AWS for *aws-nuke*.

## Using CLI Flags

To use *static credentials* the command line flags `--access-key-id` and `--secret-access-key`
are required. The flag `--session-token` is only required for temporary sessions.

`--profile` is also available if you are using the [AWS Config](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

### AWS Config

You can also authenticate using the [AWS Config](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
file. This can also have static credentials in them, but you can also use profiles. These files are generally located
at `~/.aws/config` and `~/.aws/credentials`.

To use *shared profiles* the command line flag `--profile` is required. The profile must be either defined with static
credentials in the [shared credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html) or in [shared config file](https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html) with an assuming role.

## Environment Variables

To use *static credentials* via environment variables, export `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and
optionally if using a temporary session `AWS_SESSION_TOKEN`.

