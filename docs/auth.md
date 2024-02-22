# Authentication

The authentication for aws-nuke is a bit custom but still done through the AWS SDK. In a future version, we will 
likely switch to using the AWS SDK directly to handle authentication.

## CLI Flags

The following flags are available for authentication:

- `--access-key-id` - The AWS access key ID
- `--secret-access-key` - The AWS secret access key
- `--session-token` - The AWS session token
- `--profile` - The AWS profile to use
- `--region` - The AWS region to use
- `--assume-role` - The ARN of the role to assume
- `--assume-role-session-name` - The session name to use when assuming a role
- `--assume-role-external-id` - The external ID to use when assuming a role

### Static Credentials (CLI)

To use *static credentials* the command line flags `--access-key-id` and `--secret-access-key`
are required. The flag `--session-token` is only required for temporary sessions provided to you by the AWS STS service.

**Note:** this is mutually exclusive with `--profile`.

### Static Credentials (Profiles)

`--profile` is also available if you are using the [AWS Config](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

**Note:** this is mutually exclusive with `--access-key-id` and `--secret-access-key`.

#### AWS Config

You can also authenticate using the [AWS Config](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)
file. This can also have static credentials in them, but you can also use profiles. These files are generally located
at `~/.aws/config` and `~/.aws/credentials`.

To use *shared profiles* the command line flag `--profile` is required. The profile must be either defined with static
credentials in the [shared credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html) or in [shared config file](https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html) with an assuming role.

## Environment Variables

The following environment variables are available for authentication:

- `AWS_ACCESS_KEY_ID` - The AWS access key ID
- `AWS_SECRET_ACCESS_KEY` - The AWS secret access key
- `AWS_SESSION_TOKEN` - The AWS session token
- `AWS_PROFILE` - The AWS profile to use
- `AWS_REGION` - The AWS region to use
- `AWS_ASSUME_ROLE` - The ARN of the role to assume
- `AWS_ASSUME_ROLE_SESSION_NAME` - The session name to use when assuming a role
- `AWS_ASSUME_ROLE_EXTERNAL_ID` - The external ID to use when assuming a role
