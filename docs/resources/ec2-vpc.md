---
generated: true
---

# EC2VPC


## Resource

```text
EC2VPC
```

### Alternative Resource

!!! note - Cloud Control API - Alternative Resource
    This resource can also be controlled and used via Cloud Control API. Please refer to the documentation for
    [Cloud Control Resources](../config-cloud-control.md) for more information.

```text
AWS::EC2::VPC
```


!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [EC2Subnet](./ec2-subnet.md)
- [EC2RouteTable](./ec2-route-table.md)
- [EC2DHCPOption](./ec2-dhcp-option.md)
- [EC2NetworkACL](./ec2-network-acl.md)
- [EC2NetworkInterface](./ec2-network-interface.md)
- [EC2InternetGatewayAttachment](./ec2-internet-gateway-attachment.md)
- [EC2VPCEndpoint](./ec2-vpc-endpoint.md)
- [EC2VPCPeeringConnection](./ec2-vpc-peering-connection.md)
- [EC2VPNGateway](./ec2-vpn-gateway.md)
- [EC2EgressOnlyInternetGateway](./ec2-egress-only-internet-gateway.md)

## Deprecated Aliases

!!! warning
    This resource has deprecated aliases associated with it. Deprecated Aliases will be removed in the next major
    release of aws-nuke. Please update your configuration to use the new resource name.

- `EC2Vpc`