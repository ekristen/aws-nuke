# IAM Role

This will remove all IAM Roles an AWS account.

## Settings

- `IncludeServiceLinkedRoles`

### IncludeServiceLinkedRoles

By default, service linked roles are excluded from the deletion process. This setting allows you to include them in the
deletion process now that AWS allows for them to be removed.

Default is `false`.
