# S3 Bucket

This will remove all S3 buckets from an AWS account. The following actions are performed, some with control settings.

- Remove Bucket Policy
- Remove Logging Configuration
- Remove All Legal Holds
  - This only happens if the `RemoveObjectLegalHold` setting is set to `true`
- Remove All Versions
  - This will include bypassing any Object Lock governance retention settings if the `BypassGovernanceRetention`
    setting is set to `true`
- Remove All Objects
  - This will include bypassing any Object Lock governance retention settings if the `BypassGovernanceRetention`
    setting is set to `true`

## Resource

```text
S3Bucket
```

## Settings

- `BypassGovernanceRetention` 
- `RemoveObjectLegalHold`

### BypassGovernanceRetention

Specifies whether an S3 Object Lock should bypass Governance-mode restrictions to process object retention configuration
changes or deletion. Default is `false`.

### BypassLegalHold

!!! warning
    This will result in additional S3 API calls. The entire bucket has to be enumerated for all the objects, every
    object then is the result of an API call to remove the legal hold. Regardless if a legal hold existed or not. This
    is because it would require an additional API call to check if a legal hold exists on an object.

Specifies whether S3 Object Lock should remove any legal hold configuration from objects in the S3 bucket.
Default is `false`.

