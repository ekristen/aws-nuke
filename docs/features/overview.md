There are a number of new features with this version these features range from usability to debugging purposes.

Some of the new features include:

- [Global Filters](global-filters.md)
- [Run Against All Enabled Regions](enabled-regions.md)
- [Bypass Alias Check - Allow the skip of an alias on an account](bypass-alias-check.md)
- [Signed Binaries](signed-binaries.md)
- [Filter Groups (Experimental)](filter-groups.md)
- [Name Expansion](name-expansion.md)

Additionally, there are a few new sub commands to the tool to help with setup and debugging purposes:

First, there is a new `explain-account` command that will attempt to perform basic authentication against AWS with
whatever information is has available to it, and print it out to screen, this is useful for seeing which account is
actually being targeted and what credentials the tool and AWS sees as being used.
[more information](../cli-usage.md#aws-nuke-explain-account)

Second, there is a new `explain-config` command that will attempt to parse the config file and print out the information
for the configuration file, such as total resource types, filtered resources types, number of includes or excludes with
optional flags to describe in detail all the resource types that fit the various categories.
[more information](../cli-usage.md#aws-nuke-explain-config)


