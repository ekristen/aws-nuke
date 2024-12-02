---
generated: true
---

# ELBv2


## Resource

```text
ELBv2
```

## Properties


- `ARN`: ARN of the load balancer
- `CreatedTime`: Creation time of the load balancer
- `Name`: Name of the load balancer
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


### DisableDeletionProtection

!!! note
    There is currently no description for this setting. Often times settings are fairly self-explanatory. However, we
    are working on adding descriptions for all settings.

```text
DisableDeletionProtection
```

