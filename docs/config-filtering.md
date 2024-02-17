!!! warning
    Filtering is a powerful tool, but it is also a double-edged sword. It is easy to make mistakes in the filter
    configuration. Also, since aws-nuke is in continuous development, there is always a possibility to introduce new
    bugs, no matter how careful we review new code.

# Filtering

Filtering is used to exclude or include resources from being deleted. This is important for a number of reasons to
include but limited to removing the user that runs the tool.

!!! note
    Filters are `OR'd` together. This means that if a resource matches any filter, it will be excluded from deletion.
    Currently, there is no way to do `AND'ing` of filters.

## Global

Filters are traditionally done against a specific resource. However, `__global__` as been introduced as a unique
resource type that can be used to apply filters to all defined resources. It's all or nothing, global cannot be used to
against some resources and not others.

Global works by taking all filters defined under `__global__` and prepends to any filters found for a resource type. If
a resource does NOT have any filters defined, the `__global__` ones will still be used.

### Example

In this example, we are ignoring all resources that have the tag `aws-nuke` set to `ignore`. Additionally filtering
a specific instance by its `id`. When the `EC2Instance` resource is processed, it will have both filters applied. These

```yaml
__global__:
  - property: tag:aws-nuke
    value: "ignore"

EC2Instance:
  - "i-01b489457a60298dd"
```

This will ultimately render as the following filters for the `EC2Instance` resource:

```yaml
EC2Instance:
  - "i-01b489457a60298dd"
  - property: tag:aws-nuke
    value: "ignore"
```

## Types

The following are comparisons  that you can use to filter resources. These are used in the configuration file.

- `exact`
- `contains`
- `glob`
- `regex` 
- `dateOlderThan`

To use a non-default comparison type, it is required to specify an object with `type` and `value` instead of the
plain string.

These types can be used to simplify the configuration. For example, it is possible to protect all access keys of a
single user by using `glob`:

```yaml
IAMUserAccessKey:
- type: glob
  value: "admin -> *"
```

### Exact

The identifier must exactly match the given string. **This is the default.**

Exact is just that, an exact match to a resource. The following examples are identical for the `exact` filter.

```yaml
IAMUser:
- AWSNukeUser
- type: exact
  value: AWSNukeUser
```

### Contains

The `contains` filter is a simple string contains match. The following examples are identical for the `contains` filter.

```yaml
IAMUser:
  - type: contains
    value: Nuke
```

### Glob

The identifier must match against the given [glob pattern](https://en.wikipedia.org/wiki/Glob_(programming)). This means the string might contain
wildcards like `*` and `?`. Note that globbing is designed for file paths, so the wildcards do not match the directory
separator (`/`). Details about the glob pattern can be found in the [library documentation](https://godoc.org/github.com/mb0/glob)

```yaml
IAMUser:
  - type: glob
    value: "AWSNuke*"
```

### Regex

The identifier must match against the given regular expression. Details about the syntax can be found
in the [library documentation](https://golang.org/pkg/regexp/syntax/).

```yaml
IAMUser:
  - type: regex
    value: "AWSNuke.*"
```

### DateOlderThan

This works by parsing the specified property into a timestamp and comparing it to the current time minus the specified
duration. The duration is specified in the `value` field. The duration syntax is based on golang's duration syntax.

> ParseDuration parses a duration string. A duration string is a possibly signed sequence of decimal numbers, each with
> optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"),
> "ms", "s", "m", "h".

Full details on duration syntax can be found in the [time library documentation](https://golang.org/pkg/time/#ParseDuration).

The value from the property is parsed as a timestamp and the following are the supported formats:

- `2006-01-02`
- `2006/01/02`
- `2006-01-02T15:04:05Z`
- `2006-01-02T15:04:05.999999999Z07:00`
- `2006-01-02T15:04:05Z07:00`

In the follow example we are filtering EC2 Images that have a `CreationDate` older than 1 hour.

```yaml
EC2Image:
  - type: dateOlderThan
    property: CreationDate
    value: 1h
```

## Properties

By default, when writing a filter if you do not specify a property, it will use the `Name` property. However, resources
that do no support Properties, aws-nuke will fall back to what is called the `Legacy String`, it's essentially a
function that returns a string representation of the resource. 

Some resources support filtering via properties. When a resource support these properties, they will be listed in
the output like in this example:

```log
global - IAMUserPolicyAttachment - 'admin -> AdministratorAccess' - [RoleName: "admin", PolicyArn: "arn:aws:iam::aws:policy/AdministratorAccess", PolicyName: "AdministratorAccess"] - would remove
```

To use properties, it is required to specify an object with `properties` and `value` instead of the plain string.

These types can be used to simplify the configuration. For example, it is possible to protect all access keys
of a single user:

```yaml
IAMUserAccessKey:
  - property: UserName
    value: "admin"
```

## Inverting

Any filter result can be inverted by using `invert: true`, for example:

```yaml
CloudFormationStack:
  - property: Name
    value: "foo"
    invert: true
```

In this case *any* CloudFormationStack ***but*** the ones called "foo" will be filtered. Be aware that *aws-nuke*
internally takes every resource and applies every filter on it. If a filter matches, it marks the node as filtered.

## Example

It is also possible to use Filter Properties and Filter Types together. For example to protect all Hosted Zone of a
specific TLD:

```yaml
Route53HostedZone:
  - property: Name
    type: glob
    value: "*.rebuy.cloud."
```

## Account Level

It is possible to filter this is important for not deleting the current user for example or for resources like S3
Buckets which have a globally shared namespace and might be hard to recreate. Currently, the filtering is based on
the resource identifier. The identifier will be printed as the first step of *aws-nuke* (eg `i-01b489457a60298dd` 
for an EC2 instance).

!!! warning
    **Even with filters you should not run aws-nuke on any AWS account, where you cannot afford to lose all resources.
    It is easy to make mistakes in the filter configuration. Also, since aws-nuke is in continuous development, there is
    always a possibility to introduce new bugs, no matter how careful we review new code.**

The filters are part of the account-specific configuration and are grouped by resource types. This is an example of a
config that deletes all resources but the `admin` user with its access permissions and two access keys:

```yaml
---
regions:
  - global
  - us-east-1

account-blocklist:
  - 1234567890

accounts:
  0987654321:
    filters:
      IAMUser:
        - "admin"
      IAMUserPolicyAttachment:
        - "admin -> AdministratorAccess"
      IAMUserAccessKey:
        - "admin -> AKSDAFRETERSDF"
        - "admin -> AFGDSGRTEWSFEY"
```

Any resource whose resource identifier exactly matches any of the filters in the list will be skipped. These will
be marked as "filtered by config" on the *aws-nuke* run.


## Presets

It might be the case that some filters are the same across multiple accounts.
This especially could happen, if provisioning tools like Terraform are used or
if IAM resources follow the same pattern.

For this case *aws-nuke* supports presets of filters, that can applied on
multiple accounts. A configuration could look like this:

```yaml
---
regions:
  - "global"
  - "eu-west-1"

account-blocklist:
  - 1234567890

accounts:
  555421337:
    presets:
      - "common"
  555133742:
    presets:
      - "common"
      - "terraform"
  555134237:
    presets:
      - "common"
      - "terraform"
    filters:
      EC2KeyPair:
        - "notebook"

presets:
  terraform:
    filters:
      S3Bucket:
        - type: glob
          value: "my-statebucket-*"
      DynamoDBTable:
        - "terraform-lock"
  common:
    filters:
      IAMRole:
        - "OrganizationAccountAccessRole"
```

## Included and Excluding

*aws-nuke* deletes a lot of resources and there might be added more at any release. Eventually, every resource should
get deleted. You might want to restrict which resources to delete. There are multiple ways to configure this.

One way are filters, which already got mentioned. This requires to know the identifier of each resource. It is also
possible to prevent whole resource types (eg `S3Bucket`) from getting deleted with two methods.

It is also possible to configure the resource types in the config file like in these examples:

```yaml
regions:
  - "us-east-1"

account-blocklist:
  - 1234567890

resource-types:
  # Specifying this in the configuration will ensure that only these three
  # resources are targeted by aws-nuke during it's run.
  targets:
    - S3Object
    - S3Bucket
    - IAMRole

accounts:
  555133742: {}
```

```yaml
regions:
  - "us-east-1"

account-blocklist:
  - 1234567890

resource-types:
  # Specifying this in the configuration will ensure that these resources
  # will be specifically excluded from aws-nuke during it's run.
  excludes:
  - IAMUser

accounts:
  555133742: {}
```

If targets are specified in multiple places (e.g. CLI and account specific), then a resource type must be specified in
all places. In other words each configuration limits the previous ones.

If an exclude is used, then all its resource types will not be deleted.

**Hint:** You can see all available resource types with this command:

```bash
aws-nuke resource-types
```

It is also possible to include and exclude resources using the command line arguments:

- The `--target` flag limits nuking to the specified resource types.
- The `--exclude` flag prevent nuking of the specified resource types.