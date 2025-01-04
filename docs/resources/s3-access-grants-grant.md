---
generated: true
---

# S3AccessGrantsGrant


## Resource

```text
S3AccessGrantsGrant
```

## Properties


- `CreatedAt`: The date and time the access grant was created.
- `GrantScope`: The scope of the access grant.
- `GranteeID`: The ARN of the grantee.
- `GranteeType`: The type of the grantee, (e.g. IAM).
- `ID`: The ID of the access grant.

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

