---
generated: true
---

# SSMAssociation


## Resource

```text
SSMAssociation
```

## Properties


- `AssociationId`: The ID of the SSM association
- `AssociationName`: The optional name of the SSM association
- `DocumentName`: The name of the SSM document applied by the association
- `InstanceId`: The managed node ID for a legacy association
- `TargetMaps`: The association target maps in key=value form
- `Targets`: The association targets in key=value form
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
