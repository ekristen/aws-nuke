---
generated: true
---

# EC2Snapshot


## Resource

```text
EC2Snapshot
```

## Properties


- `DataEncryptionKeyID`: The data encryption key identifier for the snapshot
- `Description`: The description for the snapshot
- `Encrypted`: Indicates whether the snapshot is encrypted
- `KmsKeyID`: The Amazon Resource Name (ARN) of the AWS KMS key used for encryption
- `OwnerAlias`: The AWS owner alias
- `OwnerID`: The AWS account ID of the EBS snapshot owner
- `Progress`: The progress of the snapshot as a percentage
- `RestoreExpiryTime`: Only for archived snapshots that are temporarily restored
- `SnapshotID`: The ID of the snapshot
- `StartTime`: The time stamp when the snapshot was initiated
- `State`: The snapshot state
- `StateMessage`: Encrypted Amazon EBS snapshots are copied asynchronously
- `StorageTier`: The storage tier in which the snapshot is stored
- `VolumeID`: The ID of the volume that was used to create the snapshot
- `VolumeSize`: The size of the volume in GiB
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

