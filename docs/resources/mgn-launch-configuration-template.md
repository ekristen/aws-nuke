---
generated: true
---

# MGNLaunchConfigurationTemplate

AWS Application Migration Service (MGN) Launch Configuration Template defines the configuration settings for launching target instances during the migration process. This template specifies EC2 instance settings, networking configuration, and other launch parameters.

## Resource

```text
MGNLaunchConfigurationTemplate
```

## Properties

- `LaunchConfigurationTemplateID` - The unique identifier of the launch configuration template
- `Arn` - The ARN of the launch configuration template
- `Ec2LaunchTemplateID` - The ID of the associated EC2 launch template
- `LaunchDisposition` - The launch disposition (STOPPED, STARTED)
- `TargetInstanceTypeRightSizingMethod` - The method for right-sizing the target instance type
- `CopyPrivateIp` - Whether to copy the private IP address
- `CopyTags` - Whether to copy tags to the launched instance
- `EnableMapAutoTagging` - Whether to enable automatic tagging
- `Tags` - The tags associated with the template

## Deletion Process

MGN Launch Configuration Templates are deleted directly using the `DeleteLaunchConfigurationTemplate` API call. This removes the template configuration from AWS MGN.



