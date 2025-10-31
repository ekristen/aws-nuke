---
generated: true
---

# EMRServerlessApplication


## Resource

```text
EMRServerlessApplication
```

## Properties


- `ARN`: The Amazon Resource Name (ARN) of the application
- `Architecture`: The CPU architecture of the application (ARM64 or X86_64)
- `CreatedAt`: The date and time when the application was created
- `ID`: The unique identifier of the application
- `Name`: The name of the application
- `ReleaseLabel`: The EMR release version used by the application
- `State`: The current state of the application (CREATING, CREATED, STARTING, STARTED, STOPPING, STOPPED, TERMINATED)
- `Type`: The type of application (Spark or Hive)
- `UpdatedAt`: The date and time when the application was last updated
- `tag:<key>`: Tags associated with the application

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

## Deletion Behavior

When deleting an EMR Serverless application, the resource handler will:

1. **Stop Application**: If the application is in a `STARTED` or `STARTING` state, it will be stopped first
2. **Wait for Stop**: Poll the application state until it reaches `STOPPED` or `CREATED` state
3. **Delete Application**: Once in a valid state, delete the application

!!! important
    Before deleting an application, you must first cancel or complete all running job runs.


### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [EMRServerlessJobRun](emr-serverless-job-run.md) - Manages EMR Serverless job runs (should be deleted before applications)
