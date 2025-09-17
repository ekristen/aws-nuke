---
generated: true
---

# MGNApplication

AWS Application Migration Service (MGN) Application represents a logical grouping of source servers in AWS MGN. Applications help organize and manage collections of servers that work together as part of a business application or workload.

## Resource

```text
MGNApplication
```

## Properties

- `ApplicationID` - The unique identifier of the application
- `Arn` - The ARN of the application
- `Name` - The name of the application
- `Description` - The description of the application
- `IsArchived` - Whether the application is archived
- `CreationDateTime` - The date and time the application was created
- `LastModifiedDateTime` - The date and time the application was last modified
- `Tags` - The tags associated with the application

## Deletion Process

MGN Applications are deleted directly using the `DeleteApplication` API call. This removes the application grouping from AWS MGN.


