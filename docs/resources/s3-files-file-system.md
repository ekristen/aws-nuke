---
generated: true
---

# S3FilesFileSystem


## Resource

```text
S3FilesFileSystem
```

## Properties


- `ID`: The ID of the S3 file system
- `Name`: The name of the S3 file system

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

- [S3FilesMountTarget](./s3-files-mount-target.md)
- [S3FilesAccessPoint](./s3-files-access-point.md)

