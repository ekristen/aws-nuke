# Config Contributions

## Community Presets

These are a collection of presets from the community. 

!!! warning
    These presets are built from feedback from the community, they are not routinely tested. Use at your own risk.

### Filter SSO Resources

This is a preset to filter out AWS SSO resources.

```yaml
presets:
  sso:
    filters:
      IAMSAMLProvider:
        - type: "regex"
          value: "AWSSSO_.*_DO_NOT_DELETE"
      IAMRole:
        - type: "glob"
          value: "AWSReservedSSO_*"
      IAMRolePolicyAttachment:
        - type: "glob"
          value: "AWSReservedSSO_*"
```

### Filter Control Tower

This is a preset to filter out AWS Control Tower resources.

```yaml
presets:
  controltower:
    filters:
      CloudTrailTrail:
        - type: "contains"
          value: "aws-controltower"
      CloudWatchEventsRule:
        - type: "contains"
          value: "aws-controltower"
        - property: "Name"
          type: glob
          value: "AWSControlTower*"
      EC2VPCEndpoint:
        - type: "contains"
          value: "aws-controltower"
      EC2VPC:
        - type: "contains"
          value: "aws-controltower"
      OpsWorksUserProfile:
        - type: "contains"
          value: "AWSControlTowerExecution"
      CloudWatchLogsLogGroup:
        - type: "contains"
          value: "aws-controltower"
        - type: "contains"
          value: "AWSControlTowerBP"
      CloudWatchEventsTarget:
        - type: "contains"
          value: "aws-controltower"
        - type: "glob"
          value: "Rule: AWSControlTower*"
      SNSSubscription:
        - type: "contains"
          value: "aws-controltower"
      SNSTopic:
        - type: "contains"
          value: "aws-controltower"
      EC2Subnet:
        - type: "contains"
          value: "aws-controltower"
      ConfigServiceDeliveryChannel:
        - type: "contains"
          value: "aws-controltower"
      ConfigServiceConfigurationRecorder:
        - type: "contains"
          value: "aws-controltower"
      CloudFormationStack:
        - type: "contains"
          value: "AWSControlTower"
      EC2RouteTable:
        - type: "contains"
          value: "aws-controltower"
      LambdaFunction:
        - type: "contains"
          value: "aws-controltower"
      EC2DHCPOption:
        - type: "contains"
          value: "aws-controltower"
      IAMRole:
        - type: "contains"
          value: "aws-controltower"
        - type: "contains"
          value: "AWSControlTower"
      IAMRolePolicyAttachment:
        - type: "contains"
          value: "aws-controltower"
        - type: "contains"
          value: "AWSControlTower"
      IAMRolePolicy:
        - type: "contains"
          value: "aws-controltower"
        - type: glob
          value: "AWSReservedSSO_*"
```