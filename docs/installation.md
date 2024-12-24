# Install

Preferred installation order is the following:

1. [GitHub Release](#github-releases-preferred)
2. [ekristen's homebrew tap](#ekristens-homebrew-tap-macoslinux)
3. [Homebrew Core](#homebrew-core-macoslinux)

Docker images are also available via the GitHub Container Registry.

## GitHub Releases (preferred)

!!! success - "Recommended"
    This supports all operating systems and most architectures.

You can download pre-compiled binaries from the [releases](https://github.com/ekristen/aws-nuke/releases) page.

You can use this method to retrieve the latest available version (example for the Linux-amd64 release):

```console
wget -q -O aws-nuke-latest-linux-amd64.tar.gz $(wget -q -O - 'https://api.github.com/repos/ekristen/aws-nuke/releases/latest' | jq -r '.assets[] | select(.name | match ("linux-amd64.tar.gz$")) | .browser_download_url')
```


## ekristen's Homebrew Tap (MacOS/Linux)

!!! info
    I control this tap, and it sources the binaries directly from the GitHub releases. However, it only supports MacOS.

```console
brew install ekristen/tap/aws-nuke
```

## Homebrew Core (MacOS/Linux)

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

