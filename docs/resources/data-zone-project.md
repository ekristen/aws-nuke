---
generated: true
---

# DataZoneProject


## Resource

```text
DataZoneProject
```

## Properties


- `CreatedAt`: The date and time when the project was created
- `CreatedBy`: The user who created the project
- `Description`: The description of the project
- `DomainID`: The ID of the domain that contains the project
- `DomainName`: The name of the domain that contains the project
- `ID`: The ID of the project
- `Name`: The name of the project
- `ProjectStatus`: The status of the project

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### DependsOn

!!! important - Experimental Feature
    This resource depends on a resource using the experimental feature. This means that the resource will
    only be deleted if all the resources of a particular type are deleted first or reach a terminal state.

- [DataZoneSubscription](./data-zone-subscription.md)

