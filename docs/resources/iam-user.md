---
generated: true
---

# IAMUser


## Resource

```text
IAMUser
```

## Properties


- `CreateDate`: No Description
- `HasPermissionBoundary`: No Description
- `Name`: No Description
- `PasswordLastUsed`: No Description
- `Path`: No Description
- `PermissionBoundaryARN`: No Description
- `PermissionBoundaryType`: No Description
- `UserID`: No Description
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

- `IgnorePermissionBoundary`


### IgnorePermissionBoundary

!!! note
    There is currently no description for this setting. Often times settings are fairly self-explanatory. However, we
    are working on adding descriptions for all settings.

```text
IgnorePermissionBoundary
```

### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [IAMUserAccessKey](./iam-user-access-key.md)
- [IAMUserHTTPSGitCredential](./iam-user-https-git-credential.md)
- [IAMUserGroupAttachment](./iam-user-group-attachment.md)
- [IAMUserPolicyAttachment](./iam-user-policy-attachment.md)
- [IAMVirtualMFADevice](./iam-virtual-mfa-device.md)

## Deprecated Aliases

!!! warning
    This resource has deprecated aliases associated with it. Deprecated Aliases will be removed in the next major
    release of aws-nuke. Please update your configuration to use the new resource name.

- `IamUser`