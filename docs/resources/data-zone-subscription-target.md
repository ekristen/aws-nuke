---
generated: true
---

# DataZoneSubscriptionTarget


## Resource

```text
DataZoneSubscriptionTarget
```

## Properties


- `CreatedAt`: The date and time when the subscription target was created
- `DomainID`: The ID of the domain that contains the subscription target
- `DomainName`: The name of the domain that contains the subscription target
- `EnvironmentID`: The ID of the environment that contains the subscription target
- `ID`: The ID of the subscription target
- `Name`: The name of the subscription target
- `ProjectID`: The ID of the project that contains the subscription target
- `Provider`: The provider of the subscription target
- `Type`: The type of the subscription target

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [DataZoneSubscriptionGrant](./data-zone-subscription-grant.md)

