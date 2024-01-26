# Experimental Features

## Overview

These are the experimental features hidden behind feature flags that are currently available in aws-nuke. They are all
disabled by default. These are switches that changes the actual behavior of the tool itself. Changing the behavior of
a resource is done via resource settings.

!!! note
    The original tool had configuration options called `feature-flags` which were used to enable/disable certain
    behaviors with resources, those are now called settings and `feature-flags` have been deprecated in the config.

## Usage

```console
aws-nuke run --feature-flag "wait-on-dependencies"
```

**Note:** other CLI arguments are omitted for brevity.

## Available Feature Flags

- `wait-on-dependencies` - This feature flag will cause aws-nuke to wait for all resource type dependencies to be 
  deleted before deleting the next resource type.

### wait-on-dependencies

This feature flag will cause aws-nuke to wait for all resource type dependencies to be deleted before deleting the next
resource type. This is useful for resources that have dependencies on other resources. For example, an IAM Role that has
an attached policy.

The problem is that if you delete the IAM Role first, it will fail because it has a dependency on the policy.

This feature flag will cause aws-nuke to wait for all resources of a given type to be deleted before deleting the next
resource type. This will reduce the number of errors and unnecessary API calls.