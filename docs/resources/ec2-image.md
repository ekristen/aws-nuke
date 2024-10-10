# EC2 Image

This will remove all IAM Roles an AWS account.

## Resource

```text
EC2Image
```

## Settings

- `IncludeDisabled`
- `IncludeDeprecated`
- `DisableDeregistrationProtection`

### IncludeDisabled

This will include any EC2 Images (AMI) that are disabled in the deletion process. By default, disabled images are excluded
from the discovery process.

Default is `false`.

### IncludeDeprecated

This will include any EC2 Images (AMI) that are deprecated in the deletion process. By default, deprecated images are excluded
from the discovery process.

Default is `false`.

### DisableDeregistrationProtection

This will disable the deregistration protection on the EC2 Image (AMI) prior to deletion. By default, deregistration protection
is not disabled.

Default is `false`.
