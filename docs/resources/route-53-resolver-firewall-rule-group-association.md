---
generated: true
---

# Route53ResolverFirewallRuleGroupAssociation


## Resource

```text
Route53ResolverFirewallRuleGroupAssociation
```

## Properties


- `Arn`: The ARN of the firewall rule group association
- `CreationTime`: The time the association was created (Unix time, UTC)
- `CreatorRequestID`: The unique ID for the request that created the association
- `FirewallRuleGroupID`: The ID of the associated firewall rule group
- `ID`: The ID of the firewall rule group association
- `ManagedOwnerName`: The owner of the association, if not managed by you
- `ModificationTime`: The time the association was last changed (Unix time, UTC)
- `MutationProtection`: Whether mutation protection is enabled for the association
- `Name`: The name of the firewall rule group association
- `Priority`: The processing order of the rule group within the VPC
- `Status`: The current status of the association
- `VpcID`: The ID of the associated VPC

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

## Settings

- `DisableDeletionProtection`


### DisableDeletionProtection

Disables mutation protection for the Firewall Rule Group/VPC association so that a protected association can be deleted.


```text
DisableDeletionProtection
```

