---
generated: true
---

# MGNReplicationConfigurationTemplate


## Resource

```text
MGNReplicationConfigurationTemplate
```

## Properties


- `Arn`: The ARN of the replication configuration template
- `AssociateDefaultSecurityGroup`: Whether to associate the default security group
- `BandwidthThrottling`: The bandwidth throttling setting
- `CreatePublicIP`: Whether to create a public IP
- `DataPlaneRouting`: The data plane routing setting
- `DefaultLargeStagingDiskType`: The default large staging disk type
- `EbsEncryption`: The EBS encryption setting
- `EbsEncryptionKeyArn`: The ARN of the EBS encryption key
- `ReplicationConfigurationTemplateID`: The unique identifier of the replication configuration template
- `ReplicationServerInstanceType`: The instance type for the replication server
- `StagingAreaSubnetId`: The subnet ID for the staging area
- `UseDedicatedReplicationServer`: Whether to use a dedicated replication server
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

