# aws-nuke

[![license](https://img.shields.io/github/license/ekristen/aws-nuke.svg)](https://github.com/ekristen/aws-nuke/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/ekristen/aws-nuke.svg)](https://github.com/ekristen/aws-nuke/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/aws-nuke)](https://goreportcard.com/report/github.com/ekristen/aws-nuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/bf05fb12c69f1ea7f257/maintainability)](https://codeclimate.com/github/ekristen/aws-nuke/maintainability)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/ekristen/aws-nuke/total)
![GitHub Downloads (all assets, latest release)](https://img.shields.io/github/downloads/ekristen/aws-nuke/latest/total)



## Overview

Remove all resources from an AWS account.

*aws-nuke* is stable, but it is likely that not all AWS resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/aws-nuke/issues/new).

## What's New in Version 3

Version 3 is a rewrite of this tool using [libnuke](https://github.com/ekristen/libnuke) with a focus on improving a number of the outstanding things
that I couldn't get done with the original project without separating out the core code into a library. See Goals
below for more.

This is not a comprehensive list, but here are some of the highlights:

* New Feature: Signed Darwin Binaries for macOS
* New Feature: Published Homebrew Tap (ekristen/tap/aws-nuke@3)
* New Feature: Global Filters
* New Feature: Run Against All Enabled Regions
* New Feature: Explain Account and Explain Config Commands
* Upcoming Feature: Filter Groups (**in progress**)
* Breaking Change: `root` command no longer triggers the run, must use subcommand `run` (alias: `nuke`)
* Breaking Change: CloudFormation Stacks now support a hold and wait for parent deletion process
* Breaking Change: Nested CloudFormation Stacks are now eligible for deletion and no longer omitted
* Completely rewrote the core of the tool as a dedicated library [libnuke](https://github.com/ekristen/libnuke)
  * This library has over 95% test coverage which makes iteration and new features easier to implement.
* Semantic Releases with notifications on issues / pull requests
* Context is passed throughout the entire library now, including the listing function and the removal function
  * This is in preparation for supporting AWS SDK Go v2
* New Resources
* Broke away from rebuy-de/aws-nuke project as a fork for reasons outlined in the history section

### Goals

- [x] Easier maintainability and bug fixing, see go report and code climate badges above
- [x] Adding additional tests around the core library
- [ ] Adding more tests around specific resource types
- [x] Adding additional resources and tooling to make adding resources easier
- [x] Adding documentation for adding resources and using the tool
- [ ] Consider adding DAG for dependencies between resource types and individual resources
- [ ] Support for AWS SDK Go v2

## Documentation

All documentation is in the [docs/](docs) directory and is built using [Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/). 

It is hosted at [https://ekristen.github.io/aws-nuke/](https://ekristen.github.io/aws-nuke/).

## History of this Fork

**Important:** this is a full fork of the original tool written by the folks over at [rebuy-de](https://github.com/rebuy-de).
This fork became necessary after attempting to make contributions and respond to issues to learn that the current
maintainers only have time to work on the project about once a month and while receptive to bringing in other 
people to help maintain, made it clear it would take time. Considering the feedback cycle was already weeks on 
initial communications, I had to make the hard decision to fork and maintain it.

### libnuke

I also needed a version of this tool for Azure and GCP, and initially I just copied and altered the code I needed for
Azure, but I didn't want to have to maintain multiple copies of the same code, so I decided to create 
[libnuke](https://github.com/ekristen/libnuke) to abstract all the code that was common between the two tools and write
proper unit tests for it. 

## Attribution, License, and Copyright

The rewrite of this tool to use [libnuke](https://github.com/ekristen/libnuke) would not have been possible without the
hard work that came before me on the original tool by the team and contributors over at [rebuy-de](https://github.com/rebuy-de)
and their original work on [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

This tool is licensed under the MIT license. See the [LICENSE](LICENSE) file for more information. The bulk of this
tool was rewritten to use [libnuke](https://github.com/ekristen/libnuke) which was in part originally sourced from
[rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke).

## Contribute

You can contribute to *aws-nuke* by forking this repository, making your changes and creating a Pull Request against
this repository. If you are unsure how to solve a problem or have other questions about a contributions, please create
a GitHub issue.

