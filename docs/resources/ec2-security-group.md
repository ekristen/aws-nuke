---
generated: true
---

# EC2SecurityGroup


## Resource

```text
EC2SecurityGroup
```

## Properties


- `ID`: The ID of the security group.
- `Name`: The name of the security group.
- `OwnerID`: The ID of the AWS account that owns the security group.
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

- [ELBv2](./elbv2.md)
- [EC2DefaultSecurityGroupRule](./ec2-default-security-group-rule.md)

