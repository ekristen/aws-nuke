# Config - Cloud Control

aws-nuke supports removing resources via the AWS Cloud Control API.

There are number of Cloud Control resources that are automatically registered as resources that can be removed by 
aws-nuke. Additionally, there are a number of resources implemented in aws-nuke that have a Cloud Control equivalent,
this is called an **alternative resource**.

For the subset of Cloud Control supported resources that are registered with aws-nuke they work like any other resource,
but they are registered with their Cloud Control API name (i.e. `AWS::Bedrock::Agent`). 

However, there are resources that have already been implemented in aws-nuke that have a Cloud Control equivalent. For
these resources an **alternative resource** has been defined. They are **MUTUALLY EXCLUSIVE**, if you include the Cloud
Control resource in your config file, the native resource will be disabled. 

Furthermore, there are some Cloud Control resources that need special handling which are not yet supported by aws-nuke.

Finally, even though the subset of automatically supported Cloud Control resources is limited, you can configure
aws-nuke to make it try any additional resource. Either via command line flags of via the config file.

## Why Use Cloud Control Resources

The Cloud Control API is a standardized API that potentially allows you to nuke any resource regardless if it is defined
within aws-nuke or not. This is especially useful for new resources that are not yet supported by aws-nuke.

## Impact on Filters

Because of how Cloud Control API resources work vs native implemented resources in aws-nuke, not all properties are
available for filtering. For example, the `AWS::EC2::VPC` resource has a `VpcId` only, whereas the `EC2VPC` resource has
`VpcID`, `Tags`, `OwnerID` and more.

## Configuration

For the config file you have to add the resource to the `resource-types.alternatives` list:

!!! note
    If you are migrating from aws-nuke@v2 `cloud-control` is deprecated but still supported for backwards compatibility
    in the configuration file. The new key is `resource-types.alternatives`.

```yaml
resource-types:
  alternatives:
    - `AWS::EC2::TransitGateway
    - `AWS::EC2::VPC
```

If you want to use the command line, you have to add a `--cloud-control` flag for each resource you want to add:

!!! important
    This will not limit the resources to only these two resources, but will add them to the list of resources that are
    automatically removed via Cloud Control.

```console
aws-nuke run \
  -c nuke-config.yaml \
  --cloud-control `AWS::EC2::TransitGateway \
  --cloud-control `AWS::EC2::VPC
```

## Supported Resources

These are the resources that are automatically supported by aws-nuke directly as Cloud Control resources that are
automatically scanned.

- `AWS::AppFlow::ConnectorProfile`
- `AWS::AppFlow::Flow`
- `AWS::AppRunner::Service`
- `AWS::ApplicationInsights::Application`
- `AWS::Backup::Framework`
- `AWS::ECR::PullThroughCacheRule`
- `AWS::ECR::RegistryPolicy`
- `AWS::ECR::ReplicationConfiguration`
- `AWS::MWAA::Environment`
- `AWS::Synthetics::Canary`
- `AWS::Timestream::Database`
- `AWS::Timestream::ScheduledQuery`
- `AWS::Timestream::Table`
- `AWS::Transfer::Workflow`

## References

- [Supported Resources](https://docs.aws.amazon.com/cloudcontrolapi/latest/userguide/supported-resources.html)