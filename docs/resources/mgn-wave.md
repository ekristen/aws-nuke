---
generated: true
---

# MGNWave

AWS Application Migration Service (MGN) Wave represents a collection of applications that are migrated together as a batch. Waves help organize migration activities by grouping applications that should be migrated in sequence or at the same time.

## Resource

```text
MGNWave
```

## Properties

- `WaveID` - The unique identifier of the wave
- `Arn` - The ARN of the wave
- `Name` - The name of the wave
- `Description` - The description of the wave
- `IsArchived` - Whether the wave is archived
- `CreationDateTime` - The date and time the wave was created
- `LastModifiedDateTime` - The date and time the wave was last modified
- `Tags` - The tags associated with the wave

## Deletion Process

MGN Waves are deleted directly using the `DeleteWave` API call. This removes the wave grouping from AWS MGN.



