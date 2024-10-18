# AccessAnalyzerArchiveRule


## Resource

```text
AccessAnalyzerArchiveRule
```

## Properties

Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
names to write filters for what you want to **keep** and omit from the nuke process.


- `AnalyzerName`: The name of the analyzer the rule is associated with
- `RuleName`: The name of the archive rule

## String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

## Deprecated Aliases

!!! warning
    This resource has deprecated aliases associated with it. Please use the new resource name.

- `ArchiveRule`