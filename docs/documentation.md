## Overview

This documentation is generated using Material for MkDocs and can be run very easily locally providing you have docker
available.

All documentation resides within the `docs` directory and is written in markdown. The `mkdocs.yml` file is used to
configure the site and the `docs` directory is used to store the markdown files.

### Running Locally

```console
make docs-serve
```

## Resource Documentation

With `version 3` of aws-nuke we have introduced a new feature to allow generating documentation for resources. This
feature is still in its early stages, and we are working on adding more resources to it. If you would like to help us
with this, please feel free to contribute to the project.

Please see [Converting a resource for self documenting](resources.md#converting-a-resource-for-self-documenting) for
more information on how to properly convert an existing resource to be self documenting.

!!! note
    Not all resources can have documentation generated with this feature. It must be implemented for each resource
    individually.

### How It Works

The underlying library that drives the bulk of this tool is [libnuke](https://github.com/ekristen/libnuke). This library
has tooling to help generate documentation for a resource. Primary the library focuses on inspecting the resource struct
and generating documentation based on the fields of the struct for properties.

There's an additional tool called `generate-docs`. This command is used to generate documentation for a resource and
write it to disk. This command leverages the struct inspection to get details about the properties and intertwine them
with the markdown template to generate the documentation.

However, for this to work the resource needs to be updated to export all it's fields. This is done by capitalizing the
first letter of the field name. The field name should match what the existing property name is if it is defined.

#### Generating Documentation for All Resources

```console
go run tools/generate-docs/docs.go --write
```

#### Generating Documentation for a Single Resource

```console
go run tools/generate-docs/docs.go --resource EC2Instance --write
```
