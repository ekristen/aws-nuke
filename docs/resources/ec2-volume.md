---
generated: true
---

# EC2Volume


## Resource

```text
EC2Volume
```

## Properties


- `AvailabilityZone`: The Availability Zone in which the volume was created
- `CreateTime`: The time stamp when volume creation was initiated
- `Encrypted`: Indicates whether the volume is encrypted
- `Iops`: The number of I/O operations per second (IOPS)
- `KmsKeyID`: The Amazon Resource Name (ARN) of the AWS KMS key used for encryption
- `MultiAttachEnabled`: Indicates whether Amazon EBS Multi-Attach is enabled
- `Size`: The size of the volume in GiB
- `State`: The state of the volume (creating, available, in-use, deleting, deleted, error)
- `Throughput`: The throughput that the volume supports in MiB/s
- `VolumeID`: The ID of the EBS volume
- `VolumeType`: The volume type (gp2, gp3, io1, io2, st1, sc1, standard)
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

