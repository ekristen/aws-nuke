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
