---
generated: true
---

# MGNReplicationConfigurationTemplate

AWS Application Migration Service (MGN) Replication Configuration Template defines the settings for data replication during the migration process. This template specifies replication server configuration, networking settings, bandwidth throttling, and encryption parameters.

## Resource

```text
MGNReplicationConfigurationTemplate
```

## Properties

- `ReplicationConfigurationTemplateID` - The unique identifier of the replication configuration template
- `Arn` - The ARN of the replication configuration template
- `StagingAreaSubnetId` - The subnet ID for the staging area
- `AssociateDefaultSecurityGroup` - Whether to associate the default security group
- `BandwidthThrottling` - The bandwidth throttling setting
- `CreatePublicIP` - Whether to create a public IP
- `DataPlaneRouting` - The data plane routing setting
- `DefaultLargeStagingDiskType` - The default large staging disk type
- `EbsEncryption` - The EBS encryption setting
- `EbsEncryptionKeyArn` - The ARN of the EBS encryption key
- `ReplicationServerInstanceType` - The instance type for the replication server
- `UseDedicatedReplicationServer` - Whether to use a dedicated replication server
- `Tags` - The tags associated with the template

## Deletion Process

MGN Replication Configuration Templates are deleted directly using the `DeleteReplicationConfigurationTemplate` API call. This removes the template configuration from AWS MGN.



