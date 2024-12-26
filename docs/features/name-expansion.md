# Name Expansion

This allows you to use wildcards in the resource names to match multiple resources. This is primarily useful when you
want to target a group of resource type for either inclusion or exclusion. 

Resource Name expansion is valid for use in the following areas:

!!! warning
This feature is currently **NOT** supported in filters.

- cli includes/excludes
- config resource types includes/excludes
- account resource types includes/excludes

## Examples

### CLI

```console
aws-nuke run --config config.yaml --include "Cognito*"
```

This can also be used with `resource-types` subcommand to see what resource types are available, and you can specify
multiple wildcard arguments.

```console
aws-nuke resource-types "Cognito*" "IAM*"
```

### Config

```yaml
resource-types:
  includes:
    - "Cognito*"
  excludes:
    - "OpsWorks*"
```

### Account Config

```yaml
accounts:
  '012345678912':
    resource-types:
      includes:
        - "Cognito*"
      excludes:
        - "OpsWorks*"
```

