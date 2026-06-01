---
generated: true
---

# CloudFormationStack


## Resource

```text
CloudFormationStack
```

## Properties


- `CreationTime`: No Description
- `LastUpdatedTime`: No Description
- `Name`: No Description
- `Status`: No Description
- `tag:<key>:`: This resource has tags with property `Tags`. These are key/value pairs that are
	added as their own property with the prefix of `tag:` (e.g. [tag:example: "value"]) 

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

## Settings

- `DisableDeletionProtection`
- `CreateRoleToDeleteStack`
- `UseCurrentRoleToDeleteStack`


### DisableDeletionProtection

When enabled, aws-nuke will automatically disable termination protection on a CloudFormation stack before
attempting to delete it. Without this setting, stacks with termination protection enabled will fail to delete.

```yaml
CloudFormationStack:
  DisableDeletionProtection: "true"
```


### CreateRoleToDeleteStack

When enabled, aws-nuke will create a temporary IAM role to delete a stack whose original execution role no longer
exists or cannot be assumed. The temporary role is tagged with `Managed: aws-nuke` and is cleaned up after deletion.

```yaml
CloudFormationStack:
  CreateRoleToDeleteStack: "true"
```


### UseCurrentRoleToDeleteStack

When enabled, aws-nuke overrides the stack's associated IAM role with the caller's current role during deletion.
The caller's role ARN is resolved via STS `GetCallerIdentity` and passed as the `RoleARN` parameter on `DeleteStack`
calls. This applies to both normal deletion and `DELETE_FAILED` retry paths.

This is useful when SCPs deny actions from the stack's original creation role (e.g. CDK `cfn-exec-role`) during
account cleanup.

```yaml
CloudFormationStack:
  UseCurrentRoleToDeleteStack: "true"
```

!!! warning "Security Consideration"
    Enabling this setting may broaden the permissions available during stack deletion. The role running aws-nuke
    typically has broader permissions than the stack's original execution role. Be aware that stack deletion
    operations (such as deleting resources within the stack) will execute with the caller's role permissions
    rather than the more constrained original stack role.

!!! note "Assumed Role Requirement"
    This setting only takes effect when aws-nuke is authenticated via an IAM assumed role. If aws-nuke is running
    as an IAM user or using any other authentication method that is not an assumed role, this setting is effectively
    a no-op and stack deletion falls back to normal behavior (using the stack's original role or no role).

!!! note "IAM Path Prefix Limitation"
    If the assumed role has an IAM path prefix (e.g. `arn:aws:iam::123456789012:role/my-path/MyRole`), the STS
    assumed-role ARN omits the path component. The reconstructed role ARN will not include the path, which may
    result in an incorrect ARN. This is uncommon in typical CDK or CloudFormation use cases.


