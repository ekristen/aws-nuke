---
generated: true
---

# MGNApplication


## Resource

```text
MGNApplication
```

## Properties


- `ApplicationID`: The unique identifier of the application
- `Arn`: The ARN of the application
- `CreationDateTime`: The date and time the application was created
- `Description`: The description of the application
- `IsArchived`: Whether the application is archived
- `LastModifiedDateTime`: The date and time the application was last modified
- `Name`: The name of the application
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

