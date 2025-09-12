---
generated: true
---

# MGNJob

AWS Application Migration Service (MGN) Job represents a migration job that has been initiated within AWS MGN. Jobs can be of different types such as LAUNCH, TERMINATE, and others, and track the progress of migration operations.

## Resource

```text
MGNJob
```

## Properties

- `JobID` - The unique identifier of the job
- `Arn` - The ARN of the job
- `Type` - The type of job (LAUNCH, TERMINATE, etc.)
- `Status` - The status of the job
- `InitiatedBy` - Who initiated the job
- `CreationDateTime` - The date and time the job was created
- `EndDateTime` - The date and time the job ended
- `Tags` - The tags associated with the job

## Deletion Process

MGN Jobs are deleted directly using the `DeleteJob` API call. This removes the job record from AWS MGN.

