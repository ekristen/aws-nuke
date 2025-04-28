---
generated: true
---

# AccessAnalyzer


## Resource

```text
AccessAnalyzer
```

### Alternative Resource

!!! warning - Cloud Control API - Alternative Resource
    This resource conflicts with an alternative resource that can be controlled and used via Cloud Control API. If you
    use this alternative resource, please note that any properties listed on this page may not be valid. You will need
    run the tool to determine what properties are available for the alternative resource via the Cloud Control API.
    Please refer to the documentation for [Cloud Control Resources](../config-cloud-control.md) for more information.

```text
AWS::AccessAnalyzer::Analyzer
```
## Properties


- `ARN`: The ARN of the analyzer
- `Name`: The name of the analyzer
- `Status`: The status of the analyzer
- `Type`: The type of the analyzer
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

