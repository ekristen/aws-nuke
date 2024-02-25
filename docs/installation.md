# Install

## Install the pre-compiled binary 

### Homebrew Tap (MacOS/Linux)

```console
brew install ekristen/tap/aws-nuke@3
```

!!! warning "Brew Warning"
    `brew install aws-nuke` will install the rebuy-aws version of aws-nuke, which is not the same as this version.

## Releases

You can download pre-compiled binaries from the [releases](https://github.com/ekristen/aws-nuke/releases) page.

## Docker

Registries:

- [ghcr.io/ekristen/aws-nuke](https://github.com/ekristen/aws-nuke/pkgs/container/aws-nuke)

You can run **aws-nuke** with Docker by using a command like this:

## Source

To compile **aws-nuke** from source you need a working [Golang](https://golang.org/doc/install) development environment and [goreleaser](https://goreleaser.com/install/).

**aws-nuke** uses go modules and so the clone path should not matter. Then simply change directory into the clone and run:

```bash
goreleaser --clean --snapshot --single-target
```

