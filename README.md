# aws-nuke

[![license](https://img.shields.io/github/license/ekristen/aws-nuke.svg)](https://github.com/ekristen/aws-nuke/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/ekristen/aws-nuke.svg)](https://github.com/ekristen/aws-nuke/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/ekristen/aws-nuke)](https://goreportcard.com/report/github.com/ekristen/aws-nuke)
[![Maintainability](https://api.codeclimate.com/v1/badges/bf05fb12c69f1ea7f257/maintainability)](https://codeclimate.com/github/ekristen/aws-nuke/maintainability)

**Forked from [rebuy-de/aws-nuke](https://github.com/rebuy-de/aws-nuke)**

## Overview

Remove all resources from an AWS account.

*aws-nuke* is stable, but it is likely that not all AWS resources are covered by it. Be encouraged to add missing
resources and create a Pull Request or to create an [Issue](https://github.com/ekristen/aws-nuke/issues/new).

## Documentation

All documentation is in the [docs/](docs/) directory and is built using [MkDocs](https://www.mkdocs.org/). However,
all the documentation is hosted at [https://ekristen.github.io/aws-nuke/](https://ekristen.github.io/aws-nuke/).

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

## Version 3

Version 3 is a rewrite of this tool using [libnuke](https://github.com/ekristen/libnuke) with a focus on improving a number of the outstanding things
that I couldn't get done with the original project without separating out the core code into a library.

### Changes

- The root command will result in help now on v3, the primary nuke command moved to `nuke`.
- CloudFormation Stacks now support a hold and wait for parent deletion process.
- Nested CloudFormation Stacks are now eligible for deletion and no longer omitted.
- The entire resource lister format has changed and requires a struct.
- Context is passed throughout the entire library now, including the listing function and the removal function.

### Goals

- Adding additional tests
- Adding additional resources
- Adding Documentation for adding resources and using the tool
- Consider adding DAG for dependencies between resource types and individual resources
  - This will improve the process of deleting resources that have dependencies on other resources and reduce 
    errors and unnecessary API calls.

## Documentation

The project is built to have the documentation right alongside the code in the `docs/` directory leveraging 
[Material for Mkdocs](https://squidfunk.github.io/mkdocs-material/)

In the root of the project exists mkdocs.yml which drives the configuration for the documentation.

This README.md is currently copied to `docs/index.md` and the documentation is automatically published to the GitHub
pages location for this repository using a GitHub Action workflow. It does not use the `gh-pages` branch.

## Contribute

You can contribute to *aws-nuke* by forking this repository, making your changes and creating a Pull Request against
this repository. If you are unsure how to solve a problem or have other questions about a contributions, please create
a GitHub issue.
