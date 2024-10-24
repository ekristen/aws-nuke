# Resources Overview

This is the start of the documentation for all resources handled by aws-nuke. Eventually each resource will have its own
page with detailed information on how to use it, what settings are available, and what the resource does.

## Properties

!!! note
    Not all Resource Types within aws-nuke have properties defined. If a resource type does not have properties defined,
    then the only matching that can be done is against its String Representation.

Properties are what [Filters](../config-filtering.md) are written against in your configuration. You use the property
names to write filters for what you want to **keep** and omit from the nuke process.

## String Representation

The string representation of a resource is generally the value of the Name, ID or ARN field of the resource. Not all
resources support properties. To write a filter against the string representation, simply omit the `property` field in
the filter.

The string value is always what is used in the output of the log format when a resource is identified.