# Starter Configuration

This is a good starting configuration for `aws-nuke`. This configuration will help you get started with the tool and
give you a good idea of what you can do with it.

By default, many of the settings are populated. Many of the resources that are deprecated or not available are excluded.

Additionally, there are 3 presets for common configurations of things you might want to filter (i.e. keep around).

!!! note
    You must replace the account ID with your own account ID. This is a placeholder account ID.

!!! warning
    This does not **cover** all settings, nor does it protect against resources that you might want to keep around, this
    is a **starting configuration only**.

```yaml
regions:
  - global
  - us-east-1
  - us-east-2

blocklist:
  - "987654321098" # Production Account

settings:
  EC2Image:
    IncludeDisabled: true
    IncludeDeprecated: true
    DisableDeregistrationProtection: true
  EC2Instance:
    DisableStopProtection: true
    DisableDeletionProtection: true
  RDSInstance:
    DisableDeletionProtection: true
  CloudFormationStack:
    DisableDeletionProtection: true
  DynamoDBTable:
    DisableDeletionProtection: true

resource-types:
  excludes:
    - S3Object # Excluded because S3 bucket removal handles removing all S3Objects
    - ServiceCatalogTagOption # Excluded due to https://github.com/rebuy-de/aws-nuke/issues/515
    - ServiceCatalogTagOptionPortfolioAttachment # Excluded due to https://github.com/rebuy-de/aws-nuke/issues/515
    - FMSNotificationChannel # Excluded because it's not available
    - FMSPolicy # Excluded because it's not available
    - MachineLearningMLModel # Excluded due to ML being unavailable
    - MachineLearningDataSource # Excluded due to ML being unavailable
    - MachineLearningBranchPrediction # Excluded due to ML being unavailable
    - MachineLearningEvaluation # Excluded due to ML being unavailable
    - RoboMakerDeploymentJob # Deprecated Service
    - RoboMakerFleet # Deprecated Service
    - RoboMakerRobot # Deprecated Service
    - RoboMakerSimulationJob
    - RoboMakerRobotApplication
    - RoboMakerSimulationApplication
    - OpsWorksApp # Deprecated service
    - OpsWorksInstance # Deprecated service
    - OpsWorksLayer # Deprecated service
    - OpsWorksUserProfile # Deprecated service
    - OpsWorksCMBackup # Deprecated service
    - OpsWorksCMServer # Deprecated service
    - OpsWorksCMServerState # Deprecated service
    - CodeStarProject # Deprecated service
    - CodeStarConnection # Deprecated service
    - CodeStarNotification # Deprecated service
    - Cloud9Environment # Deprecated service
    - CloudSearchDomain # Deprecated service
    - RedshiftServerlessSnapshot # Deprecated service
    - RedshiftServerlessNamespace # Deprecated service
    - RedshiftServerlessWorkgroup # Deprecated service

presets:
  common:
    filters:
      BudgetsBudget:
        - property: Name
          value: "My Zero-Spend Budget"

  organization:
    filters:
      IAMSAMLProvider:
        - property: ARN
          type: contains
          value: "AWSSSO"
      IAMRole:
        - property: Name
          type: contains
          value: "OrganizationAccountAccessRole"
      IAMRolePolicyAttachment:
        - property: RoleName
          value: "OrganizationAccountAccessRole"

  defaults:
    filters:
      EC2Subnet:
        - property: DefaultVPC
          value: "true"
      EC2DefaultSecurityGroupRule:
        - property: DefaultVPC
          value: "true"
      EC2DHCPOption:
        - property: DefaultVPC
          value: "true"
      EC2VPC:
        - property: IsDefault
          value: "true"
      EC2InternetGateway:
        - property: DefaultVPC
          value: "true"
      EC2InternetGatewayAttachment:
        - property: DefaultVPC
          value: "true"

accounts:
  '012345678901':
    presets:
      - common
      - organization
      - defaults

```