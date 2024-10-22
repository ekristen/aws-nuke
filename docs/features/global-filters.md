# Global Filters

Global filters are filters that are applied to all resources. They are defined using a special resource name called
`__global__`. The global filters are pre-pended to all resources before any other filters for the specific resource
are applied.

!!! note
    This is a pseudo resource so to use it for filtering it can only be done in the supported filter locations, 
    such as `presets` or `accounts`.

[Full Documentation](../config-filtering.md#global)