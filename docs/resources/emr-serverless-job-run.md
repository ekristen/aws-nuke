---
generated: true
---

# EMRServerlessJobRun


## Resource

```text
EMRServerlessJobRun
```

## Properties


- `ARN`: The Amazon Resource Name (ARN) of the job run
- `ApplicationID`: The ID of the EMR Serverless application running this job
- `ApplicationName`: The name of the EMR Serverless application running this job
- `CreatedAt`: The date and time when the job run was created
- `JobRunID`: The unique identifier of the job run
- `Name`: The name of the job run
- `State`: The current state of the job run (SUBMITTED, PENDING, SCHEDULED, RUNNING, SUCCESS, FAILED, CANCELLING, CANCELLED)
- `UpdatedAt`: The date and time when the job run was last updated
- `tag:<key>`: Tags associated with the job run

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

## Deletion Behavior

When deleting an EMR Serverless job run, the resource handler will:

1. **Cancel the Job Run**: Call the CancelJobRun API to cancel the running job
2. **Filter Non-Cancellable Jobs**: Only job runs in cancellable states (SUBMITTED, PENDING, SCHEDULED, RUNNING) are included

!!! note
    Only active job runs that can be cancelled are discovered and managed by this resource. Completed, failed, or already cancelled jobs are automatically filtered out during the listing phase.

## Usage Example

To cancel all running job runs except those in production:

```yaml
EMRServerlessJobRun:
  - property: tag:Environment
    value: "production"
```

To cancel job runs for a specific application:

```yaml
EMRServerlessJobRun:
  - property: ApplicationName
    value: "my-critical-app"
```

!!! warning
    Cancelling job runs will interrupt ongoing data processing. Ensure critical jobs are protected via filters before running aws-nuke.
