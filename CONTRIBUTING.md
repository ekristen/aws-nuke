# Contributing

Thank you for wanting to contribute to *aws-nuke*.

Because of the amount of AWS services and their rate of change, we rely on your participation. For the same reason we
can only act retroactive on changes of AWS services. Otherwise, it would be a full time job to keep up with AWS.

## How Can I Contribute?

### Some Resource Is Not Supported by *aws-nuke*

If a resource is not yet supported by *aws-nuke*, you have two options to resolve this:

* File [an issue](https://github.com/rebuy-de/aws-nuke/issues/new) and describe which resource is missing. This way someone can take care of it.
* Add the resource yourself and open a Pull Request. Please follow the guidelines below to see how to create 
  such a resource.

### Some Resource Does Not Get Deleted

Please check the following points before creating a bug issue:

* Is the resource actually supported by *aws-nuke*? If not, please follow the guidelines above.
* Are there permission problems? In this case *aws-nuke* will print errors that usually contain the status code `403`.
* Did you just get scared by an error that was printed? *aws-nuke* does not know about dependencies between resources. 
  To work around this it will just retry deleting all resources in multiple iterations. Therefore, it is normal that
  there are a lot of dependency errors in the first one. The iterations are separated by lines starting with
  `Removal requested:` and only the errors in the last block indicate actual errors.

File [an issue](https://github.com/rebuy-de/aws-nuke/issues/new) and describe as accurately as possible how to generate the resource on AWS that cause the
errors in *aws-nuke*. Ideally this is provided in a reproducible way like a Terraform template or AWS CLI commands.

### I Have Ideas to Improve *aws-nuke*

You should take these steps if you have an idea how to improve *aws-nuke*:

1. Check the [issues page](https://github.com/rebuy-de/aws-nuke/issues), whether someone already had the same or a similar idea.
2. Also check the [closed issues](https://github.com/rebuy-de/aws-nuke/issues?utf8=%E2%9C%93&q=is%3Aissue), because this might have already been implemented, but not yet released. Also,
   the idea might not be viable for obvious reasons.
3. Join the discussion, if there is already a related issue. If this is not the case, open a new issue and describe
   your idea. Afterward, we can discuss this idea and form a proposal.

### I Just Have a Question

Please use [GitHub Discussions](https://github.com/ekristen/aws-nuke/discussions)

## Resource Guidelines

### Tooling

Checkout the documentation around [resources](https://ekristen.github.io/aws-nuke/resources/) as it provides resource
format and a tool to help generate the resource.

### Consider Pagination

Most AWS resources are paginated and all resources should handle that.

### Use Properties Instead of String Functions

Currently, each resource can offer two functions to describe itself, that are used by the user to identify it and by
*aws-nuke* to filter it.

The String function is deprecated:

```golang
func (r *Resource) String() string
```

The Properties function should be used instead:

```golang
func (r *Resource) Properties() types.Properties
```

**Note:** The interface for the String function is still there, because not all resources are migrated yet. Please use
the Properties function for new resources.

### Filter Resources That Cannot Get Removed

Some AWS APIs list resources, that cannot be deleted. For example:

* Resources that are already deleted, but still listed for some time (e.g. EC2 Instances)
* Resources that are created by AWS, but cannot be deleted by the user (e.g. some IAM Roles)

Those resources should be excluded in the filter step, rather than in the list step.

## Styleguide

### Go

#### golangci-lint

There is an extensive golangci-lint configuration in the repository. Please make sure to run `golangci-lint run` before
committing any changes.

#### Code Format

Like almost all Go projects, we are using `go fmt` as a single source of truth for formatting the source code. Please
use `go fmt` before committing any change.

#### Import Format

1. Standard library imports
2. Third party imports
3. AWS SDK imports
4. ekristen/libnuke imports
5. Local package imports

##### Example Import Format

```golang
package example

import (
    "context"

    "github.com/sirupsen/logrus"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	"github.com/ekristen/libnuke/pkg/settings"
	
	"github.com/ekristen/aws-nuke/pkg/types"
)
```

### Git

#### Pull Requests

We default to squash merge pull requests. This means that the commit history of the pull request will be squashed into a
single commit and then merged into the main branch. This keeps the commit history clean and easy to read.

#### Commits

We are using the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for our commit
messages. This allows us to automatically generate a changelog and version numbers.

All commits in a pull request must follow this format or the GitHub Actions will fail. 

#### Signed Commits

We require that all commits be signed.

```console
git config --global commit.gpgsign true
```

#### Setup Email

We prefer having the commit linked to the GitHub account, that is creating the Pull Request. To make this happen,
*git* must be configured with an email, that is registered with a GitHub account.

To set the email for all git commits, you can use this command:

```bash
git config --global user.email "email@example.com"
```

If you want to change the email only for the *aws-nuke* repository, you can skip the `--global` flag. You have to
make sure that you are executing this in the *aws-nuke* directory:

```bash
git config user.email "email@example.com"
```

If you already committed something with a wrong email, you can use this command:

```bash
git commit --amend --author="Author Name <email@address.com>"
```

This changes the email of the latest commit. If you have multiple commits in your branch, please squash them and
change the author afterwards.
