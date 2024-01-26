# Bypass AWS Alias Check

While we take security and precautions serious with a tool that can have devastating effects, we understand that some
users may want to skip the alias check. This is not recommended, but we understand that some users may want to do this.

This feature allows you to skip the alias check but requires additional configuration to enable.

## How it Works

There are two components to this feature:

1. The `--no-alias-check` flag (also available as `NO_ALIAS_CHECK` environment variable)
2. The `bypass-alias-check-accounts` configuration in the `config.yml` file

The account ID must be in the `bypass-alias-check-accounts` configuration for the `--no-alias-check` flag to work.

## Example Configuration

```yaml
bypass-alias-check-accounts:
  - 123456789012
```

## Example Usage

```console
aws-nuke run --config=example-config.yaml --no-alias-check
```