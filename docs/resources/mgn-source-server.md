---
generated: true
---

# MGNSourceServer

AWS Application Migration Service (MGN) Source Server represents a server that has been configured for migration using AWS MGN. Source servers are the physical or virtual machines in your source environment that you want to migrate to AWS.

## Resource

```text
MGNSourceServer
```

## Properties

- `SourceServerID` - The unique identifier of the source server
- `Arn` - The ARN of the source server
- `ReplicationType` - The type of replication (AGENT_BASED, etc.)
- `IsArchived` - Whether the source server is archived
- `LifeCycleState` - The lifecycle state of the source server
- `Hostname` - The hostname of the source server
- `FQDN` - The fully qualified domain name of the source server
- `Tags` - The tags associated with the source server

## Deletion Process

When deleting an MGN Source Server, aws-nuke performs the following steps:

1. First disconnects the source server from the MGN service using `DisconnectFromService`
2. Then deletes the source server using `DeleteSourceServer`

This ensures that replication is properly stopped before the resource is removed.

