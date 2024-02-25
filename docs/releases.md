# Releases

Releases are performed automatically based on semantic commits. The versioning is then determined from the commits.
We use a tool called [semantic-release](https://semantic-release.gitbook.io/) to determine if a release should occur
and what the version should be. This takes away the need for humans to be involved.

## Release Assets

You can find Linux, macOS and Windows binaries on the [releases page](https://github.com/ekristen/aws-nuke/releases), but we also provide containerized
versions on [ghcr.io/ekristen/aws-nuke](https://ghcr.io/ekristen/aws-nuke).

Both are available for multiple architectures (amd64, arm64 & armv7) using Docker manifests. You can reference the
main tag on any system and get the correct docker image automatically.