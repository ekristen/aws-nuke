# Install

Preferred installation order is the following:

- [Install](#install)
  - [GitHub Releases (preferred)](#github-releases-preferred)
  - [Mise](#mise)
  - [Homebrew Tap (macOS)](#homebrew-tap-macos)
  - [Homebrew Core (macOS/Linux)](#homebrew-core-macoslinux)
  - [Docker](#docker)
  - [Source](#source)
  - [Verifying Binaries](#verifying-binaries)

Docker images are also available via the GitHub Container Registry.

## GitHub Releases (preferred)

!!! success - "Recommended"
    This supports all operating systems and most architectures.

You can download pre-compiled binaries from the [releases](https://github.com/ekristen/aws-nuke/releases) page, or you can use my tool
[distillery](https://github.com/ekristen/distillery) to download and install the latest version.

```console
dist install ekristen/aws-nuke
```

## Mise

If you are an enthusiast user of [mise](https://github.com/jdx/mise), you can install it with a command like:

```console
mise use -g aws-nuke@latest
```

## Homebrew Tap (macOS)

!!! info
    I control this tap, and it sources the binaries directly from the GitHub releases. However, it only supports MacOS
    and it tends to lag a bit behind.

```console
brew install ekristen/tap/aws-nuke
```

## Homebrew Core (macOS/Linux)

!!! note
    I do not control the Homebrew Core formula, so it may not be up to date. Additionally, it is not compiled with
    goreleaser, instead it is compiled with the Homebrew build system which does not build it in the same way, for
    example it does not compile it statically.

```console
brew install aws-nuke
```
## Docker

Registries:

- [ghcr.io/ekristen/aws-nuke](https://github.com/ekristen/aws-nuke/pkgs/container/aws-nuke)

## Source

To compile **aws-nuke** from source you need a working [Golang](https://golang.org/doc/install) development environment and [goreleaser](https://goreleaser.com/install/).

**aws-nuke** uses go modules and so the clone path should not matter. Then simply change directory into the clone and run:

```bash
goreleaser build --clean --snapshot --single-target
```

## Verifying Binaries

All the binaries are signed with [cosign](https://github.com/sigstore/cosign) and are signed with keyless signatures.
You can verify the build using the public transparency log and the cosign binary.

**Note:** swap out `VERSION` with `vX.Y.Z`.

```console
cosign verify-blob \
  --signature https://github.com/ekristen/aws-nuke/releases/download/VERSION/checksums.txt.sig \
  --certificate https://github.com/ekristen/aws-nuke/releases/download/VERSION/checksums.txt.pem \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --certificate-identity "https://github.com/ekristen/aws-nuke/.github/workflows/goreleaser.yml@refs/tags/VERSION" \
  https://github.com/ekristen/aws-nuke/releases/download/VERSION/checksums.txt
```
