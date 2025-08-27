---
generated: true
---

# EC2VerifiedAccessEndpoint


## Resource

```text
EC2VerifiedAccessEndpoint
```

## Properties


- `ApplicationDomain`: The DNS name for the application (e.g., example.com)
- `AttachmentType`: The type of attachment (vpc)
- `CreationTime`: The timestamp when the Verified Access endpoint was created
- `Description`: A description for the Verified Access endpoint
- `DomainCertificateArn`: The ARN of the SSL/TLS certificate for the domain
- `EndpointType`: The type of endpoint (network-interface or load-balancer)
- `ID`: The unique identifier of the Verified Access endpoint
- `LastUpdatedTime`: The timestamp when the Verified Access endpoint was last updated
- `VerifiedAccessGroupId`: The ID of the Verified Access group this endpoint belongs to
- `tag:<key>:`: This resource has tags with property `Tags`. These are key/value pairs that are
	added as their own property with the prefix of `tag:` (e.g. [tag:example: "value"]) 

!!! note - Using Properties
    Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
    names to write filters for what you want to **keep** and omit from the nuke process.

### String Property

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.

