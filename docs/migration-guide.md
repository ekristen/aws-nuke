There's only **one** real breaking change between version 2 and version 3 and that is that the root command will
no longer trigger a run, you must use the `run` command to trigger a run.

There are a number of other changes that have been made, but they are 100% backwards compatible and warnings are provided
during run time if you are using deprecated flags or resources.

## CLI Changes

- `root` command no longer triggers the run, must use subcommand `run` (alias: `nuke`)
- `target` is deprecated, use `include` instead (on `run` command)
- `feature-flag` is a new flag used to change behavior of the tool, for resource behavior changes see [settings](config.md#settings)

## Config Changes

### New

* `settings` is a new section in the config file that allows you to change the behavior of a resource, formerly these
  were called `feature-flags` and were a top level key in the config file.

### Deprecated

- `targets` is now deprecated, use `includes` instead
- `feature-flags` is now deprecated, use `settings` instead
