---
generated: true
---

# EC2Instance


## Resource

```text
EC2Instance
```

## Properties


- `Identifier`: The instance ID (e.g. i-1234567890abcdef0)
- `ImageIdentifier`: The ID of the AMI used to launch the instance
- `InstanceState`: The current state of the instance
- `InstanceType`: The instance type (e.g. t2.micro)
- `LaunchTime`: The time the instance was launched
- `tag:&lt;key&gt;:`: This resource has tags with property `Tags`. These are key/value pairs that are
	added as their own property with the prefix of `tag:` (e.g. [tag:example: &#34;value&#34;]) 

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
- `DisableStopProtection`


### DisableDeletionProtection

!!! note
    There is currently no description for this setting. Often times settings are fairly self-explanatory. However, we
    are working on adding descriptions for all settings.

```text
DisableDeletionProtection
```


### DisableStopProtection

!!! note
    There is currently no description for this setting. Often times settings are fairly self-explanatory. However, we
    are working on adding descriptions for all settings.

```text
DisableStopProtection
```

