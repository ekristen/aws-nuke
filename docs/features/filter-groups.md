# Filter Groups

!!! important
    This feature is experimental and is disabled by default. To enable it, use the `--feature-flag "filter-groups"` CLI argument.

Filter groups allow you to filter resources based on a grouping method which allows for AND'ing filters together. By
default, all filters belong to the same group, but you can specify a group name to group filters together. 

All filters within a group are OR'd together, and all groups are AND'd together. 

[Full Documentation](../config-filtering.md#filter-groups)