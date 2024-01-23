# Configuration Migration

## Version 2.x to 3.x

The configuration file format has changed from version 2.x to 3.x. However, it is still 100% backward compatible with
the old format. The new format is more flexible and allows for more complex configurations.

### Changes

- The `targets` key has been deprecated in favor of `includes`.
- The `feature-flags` key has been deprecated in favor of `settings`.

### Migration

The migration for `targets` is very simply, simply rename the key to `includes`

```yaml
resource-types:
  targets:
    - S3Object
    - S3Bucket
    - IAMRole
```

Becomes

```yaml
resource-types:
  includes:
    - S3Object
    - S3Bucket
    - IAMRole
```

The migration for `feature-flags` takes a little more than renaming the key. The `settings` key is now used to map
settings to a specific resource and that resource's definition within the tool announces the need for a setting.

```yaml
feature-flags:
  disable-deletion-protection:
    RDSInstance: true
    EC2Instance: true
```

Becomes

```yaml
settings:
  EC2Instance:
    DisableDeletionProtection: true
  RDSInstance:
    DisableDeletionProtection: true
```