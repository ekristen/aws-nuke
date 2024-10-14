# S3Object

!!! warning
    **You should exclude this resource by default.** Not doing so can lead to deadlocks and hung runs of the tool. In 
    the next major version of aws-nuke, this resource will be excluded by default.

!!! important
    This resource is **NOT** required to remove a [S3Bucket](./s3-bucket.md). The `S3Bucket` resource will remove all
    objects in the bucket as part of the deletion process using a batch removal process.

This removes all objects from S3 buckets in an AWS account while retaining the S3 bucket itself. This resource is
useful if you want to remove a single object from a bucket, or a subset of objects without removing the entire bucket.

## Resource

```text
S3Object
```

## Settings

**No settings available.**