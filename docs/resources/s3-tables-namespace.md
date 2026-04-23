---
generated: true
---

# S3TablesNamespace


## Resource

```text
S3TablesNamespace
```

## Properties


- `CreationDate`: The date and time the namespace was created.
- `Name`: The name of the namespace.
- `TableBucketName`: The name of the table bucket the namespace belongs to.

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

- [S3TablesTable](./s3-tables-table.md)

