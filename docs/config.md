# Config

The configuration is the user supplied configuration that is used to drive the nuke process. The configuration is a YAML file that is loaded from the path specified by the --config flag.

## Sections

The configuration is broken down into the following sections:

- [blocklist](#blocklist)
- [regions](#regions)
- [accounts](#accounts)
    - [presets](#presets)
    - [filters](#filters)
    - [resource-types](#resource-types)
        - [includes](#includes)
        - [excludes](#excludes)
        - [cloud-control](#cloud-control)
        - targets (deprecated, use includes)
- [resource-types](#resource-types)
    - [includes](#includes)
    - [excludes](#excludes)
    - [cloud-control](#cloud-control)
    - targets (deprecated, use includes)
- [feature-flags](#feature-flags) (deprecated, use settings instead)
- [settings](#settings)
- [presets](#global-presets)

## Simple Example

```yaml
blocklist:
  - 1234567890

regions:
  - global
  - us-east-1

accounts:
  0987654321:
    filters:
      IAMUser:
        - "admin"
      IAMUserPolicyAttachment:
        - "admin -> AdministratorAccess"
      IAMUserAccessKey:
        - "admin -> AKSDAFRETERSDF"
        - "admin -> AFGDSGRTEWSFEY"

resource-types:
  includes:
    - IAMUser
    - IAMUserPolicyAttachment
    - IAMUserAccessKey

settings:
  EC2Instance:
    DisableDeletionProtection: true
  RDSInstance:
    DisableDeletionProtection: true
```

## Blocklist

The blocklist is simply a list of Accounts that the tool cannot run against. This is to protect the user from accidentally
running the tool against the wrong account. The blocklist must always be populated with at least one entry.

## Regions

The regions is a list of AWS regions that the tool will run against. The tool will run against all regions specified in the
configuration. If no regions are listed, then the tool will **NOT** run against any region. Regions must be explicitly
provided.

### All Enabled Regions

You may specify the special region `all` to run against all enabled regions. This will run against all regions that are
enabled in the account. It will not run against regions that are disabled. It will also automatically include the 
special region `global` which is for specific global resources.

!!! important
    The use of `all` will ignore all other regions specified in the configuration. It will only run against regions
    that are enabled in the account.

## Accounts

The accounts section is a map of AWS Account IDs to their configuration. The account ID is the key and the value is the
configuration for that account.

The configuration for each account is broken down into the following sections:

- presets
- filters
- resource-types
    - targets (deprecated, use includes)
    - includes
    - excludes
    - cloud-control

### Presets

Presets under an account entry is a list of strings that must map to a globally defined preset in the configuration.

### Filters

Filters is a map of filters against resource types. To learn more about filters, see the [Filtering](./config-filtering.md)

**Note:** filters can be defined at the account level and at the preset level.

## Resource Types

Resource types is a map of resource types to their configuration. The resource type is the key and the value is the
configuration for that resource type.

The configuration for each resource type is broken down into the following sections:

- includes
- excludes
- cloud-control
- targets (deprecated, use includes)

### Includes

Includes are a list of resource types the tool will run against. If no includes are specified, then the tool will run against
all resource types.

### Excludes

Excludes are a list of resource types the tool will not run against. If no excludes are specified, then the tool will run
against all resource types unless Includes is specified.

### Cloud Control

Cloud Control is a map of resource types to their cloud control configuration. This allows for alternative behavior when
removing resources. If a resource has a Cloud Control alternative, and you'd like to use its behavior, then you can specify
the resource type in the `cloud-control` section.

## Feature Flags

!!! warning
    Deprecated. Please use settings instead.

Feature flags are a map of resource types to their feature flag configuration. This allows for alternative behavior when
removing resources. If a resource has a feature flag alternative, and you'd like to use its behavior, then you can specify
the resource type in the `feature-flags` section.

## Settings

Settings are a map of resource types to their settings configuration. This allows for alternative behavior when removing
resources. If a resource has a setting alternative, and you'd like to use its behavior, then you can specify the resource
type in the `settings` section.

## Global Presets

To read more on global presets, see the [Presets](./config-presets.md) documentation.