---
generated: true
---

# NetworkFirewall


## Resource

```text
NetworkFirewall
```

### Alternative Resource

!!! warning - Cloud Control API - Alternative Resource
    This resource conflicts with an alternative resource that can be controlled and used via Cloud Control API. If you
    use this alternative resource, please note that any properties listed on this page may not be valid. You will need
    run the tool to determine what properties are available for the alternative resource via the Cloud Control API.
    Please refer to the documentation for [Cloud Control Resources](../config-cloud-control.md) for more information.

```text
AWS::NetworkFirewall::Firewall
```
## Properties


- `ARN`: The ARN of the firewall.
- `Name`: The name of the firewall.

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

- [NetworkFirewallLoggingConfiguration](./network-firewall-logging-configuration.md)

