# Development

## Building

The following will build the binary for the current platform and place it in the `releases` directory.

```console
goreleaser build --clean --snapshot --single-target
```

## Documentation

This is built using Material for MkDocs and can be run very easily locally providing you have docker available.

### Running Locally

```console
make docs-serve
```
