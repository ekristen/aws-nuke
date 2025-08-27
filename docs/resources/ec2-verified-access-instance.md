---
generated: true
---

# EC2VerifiedAccessInstance


## Resource

```text
EC2VerifiedAccessInstance
```

## Properties


- `CreationTime`: The timestamp when the Verified Access instance was created
- `Description`: A description for the Verified Access instance
- `ID`: The unique identifier of the Verified Access instance
- `LastUpdatedTime`: The timestamp when the Verified Access instance was last updated
- `TrustProviders`: A list of trust provider IDs attached to the Verified Access instance
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

### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [EC2VerifiedAccessGroup](./ec2-verified-access-group.md)
- [EC2VerifiedAccessEndpoint](./ec2-verified-access-endpoint.md)

