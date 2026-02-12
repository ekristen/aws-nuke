---
generated: true
---

# DataZoneSubscription


## Resource

```text
DataZoneSubscription
```

## Properties


- `CreatedAt`: The date and time when the subscription was created
- `DomainID`: The ID of the DataZone domain containing the subscription
- `DomainName`: The name of the DataZone domain containing the subscription
- `ID`: The unique identifier of the subscription
- `Status`: The current status of the subscription

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.


### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [DataZoneSubscriptionTarget](./data-zone-subscription-target.md)
