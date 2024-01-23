# AWS Nuke

Remove all resources from an AWS account.

*aws-nuke* is stable, but it is likely that not all AWS resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/aws-nuke/issues/new).

## History

This is a full fork of the original tool written by the folks over at [rebuy-de](https://github.com/rebuy-de). This fork became necessary
after attempting to make contributions and respond to issues to learn that the current maintainers only have time to
work on the project about once a month and while receptive to bringing in other people to help maintain, made it clear
it would take time. Considering the feedback cycle was already weeks on initial communications, I had to make the hard
decision to fork and maintain it.

Since then the rebuy-de team has taken up interest in responding to their issues and pull requests, but I have decided
to continue maintaining this fork as I have a few things I want to do with it that I don't think they will be interested.

## libnuke

Officially over the Christmas break of 2023, I decided to create [libnuke](https://github.com/ekristen/libnuke) which
is a library that can be used to create similar tools for other cloud providers. This library is used by both this tool,
aws-nuke, and [azure-nuke](https://github.com/ekristen/azure-nuke) and soon [gcp-nuke](https://github.com/ekristen/gcp-nuke).

I also needed a version of this tool for Azure and GCP, and initially I just copied and altered the code I needed for
Azure, but I didn't want to have to maintain multiple copies of the same code, so I decided to create
[libnuke](https://github.com/ekristen/libnuke) to abstract all the code that was common between the two tools and write proper unit tests for it. 

## Why a rewrite?

I decided to rewrite this tool for a few reasons:

- [x] I wanted to improve the build process by using `goreleaser`
- [x] I wanted to improve the release process by using `goreleaser` and publishing multi-architecture images
- [x] I also wanted to start signing all the releases
- [x] I wanted to add additional tests and improve the test coverage, more tests on the way for individual resources.
    - [libnuke](https://github.com/ekristen/libnuke) is at 94%+ overall test coverage.
- [x] I wanted to reduce the maintenance burden by abstracting the core code into a library
- [x] I wanted to make adding additional resources more easy and lowering the barrier to entry
- [x] I wanted to add a lot more documentation and examples
- [x] I wanted to take steps to make way for AWS SDK Version 2
- [ ] I wanted to add a DAG for dependencies between resource types and individual resources (this is still a work in progress)
    - This will improve the process of deleting resources that have dependencies on other resources and reduce errors and unnecessary API calls.
