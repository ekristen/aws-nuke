# Custom Endpoints

## Overview

It is possible to configure aws-nuke to run against non-default AWS endpoints. It could be used for integration testing
pointing to a local endpoint such as an S3 appliance or a Stratoscale cluster for example.

To configure aws-nuke to use custom endpoints, add the configuration directives as shown in the following example:

## Example

```yaml
regions:
  - demo10

# inspired by https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html
endpoints:
  - region: demo10
    tls_insecure_skip_verify: true
    services:
      - service: ec2
        url: https://10.16.145.115/api/v2/aws/ec2
      - service: s3
        url: https://10.16.145.115:1060
      - service: rds
        url: https://10.16.145.115/api/v2/aws/rds
      - service: elbv2
        url: https://10.16.145.115/api/v2/aws/elbv2
      - service: efs
        url: https://10.16.145.115/api/v2/aws/efs
      - service: emr
        url: https://10.16.145.115/api/v2/aws/emr
      - service: autoscaling
        url: https://10.16.145.115/api/v2/aws/autoscaling
      - service: cloudwatch
        url: https://10.16.145.115/api/v2/aws/cloudwatch
      - service: sns
        url: https://10.16.145.115/api/v2/aws/sns
      - service: iam
        url: https://10.16.145.115/api/v2/aws/iam
      - service: acm
        url: https://10.16.145.115/api/v2/aws/acm

account-blocklist:
  - "account-id-of-custom-region-prod" # production

accounts:
  "account-id-of-custom-region-demo10": {}
```

### Output

This can then be used as follows:

```log
$ aws-nuke -c config/my.yaml  --access-key-id <access-key> --secret-access-key <secret-key> --default-region demo10
aws-nuke version v2.11.0.2.gf0ad3ac.dirty - Tue Nov 26 19:15:12 IST 2019 - f0ad3aca55eb66b93b88ce2375f8ad06a7ca856f

Do you really want to nuke the account with the ID account-id-of-custom-region-demo10 and the alias 'account-id-of-custom-region-demo10'?
Do you want to continue? Enter account alias to continue.
> account-id-of-custom-region-demo10

demo10 - EC2Volume - vol-099aa1bb08454fd5bc3499897f175fd8 - [tag:Name: "volume_of_5559b38e-0a56-4078-9a6f-eb446c21cadf"] - would remove
demo10 - EC2Volume - vol-11e9b09c71924354bcb4ee77e547e7db - [tag:Name: "volume_of_e4f8c806-0235-4578-8c08-dce45d4c2952"] - would remove
demo10 - EC2Volume - vol-1a10cb3f3119451997422c435abf4275 - [tag:Name: "volume-dd2e4c4a"] - would remove
demo10 - EC2Volume - vol-1a2e649df1ef449686ef8771a078bb4e - [tag:Name: "web-server-5"] - would remove
demo10 - EC2Volume - vol-481d09bbeb334ec481c12beee6f3012e - [tag:Name: "volume_of_15b606ce-9dcd-4573-b7b1-4329bc236726"] - would remove
demo10 - EC2Volume - vol-48f6bd2bebb945848b029c80b0f2de02 - [tag:Name: "Data volume for 555e9f8a"] - would remove
demo10 - EC2Volume - vol-49f0762d84f0439da805d11b6abc1fee - [tag:Name: "Data volume for acb7f3a5"] - would remove
demo10 - EC2Volume - vol-4c34656f823542b2837ac4eaff64762b - [tag:Name: "wpdb"] - would remove
demo10 - EC2Volume - vol-875f091078134fee8d1fe3b1156a4fce - [tag:Name: "volume-f1a7c95f"] - would remove
demo10 - EC2Volume - vol-8776a0d5bd4e4aefadfa8038425edb20 - [tag:Name: "web-server-6"] - would remove
demo10 - EC2Volume - vol-8ed468bfab0b42c3bc617479b8f33600 - [tag:Name: "web-server-3"] - would remove
demo10 - EC2Volume - vol-94e0370b6ab54f03822095d74b7934b2 - [tag:Name: "web-server-2"] - would remove
demo10 - EC2Volume - vol-9ece34dfa7f64dd583ab903a1273340c - [tag:Name: "volume-4ccafc2e"] - would remove
demo10 - EC2Volume - vol-a3fb3e8800c94452aff2fcec7f06c26b - [tag:Name: "web-server-0"] - would remove
demo10 - EC2Volume - vol-a53954e17cb749a283d030f26bbaf200 - [tag:Name: "volume-5484e330"] - would remove
demo10 - EC2Volume - vol-a7afe64f4d0f4965a6703cc0cfab2ba4 - [tag:Name: "Data volume for f1a7c95f"] - would remove
demo10 - EC2Volume - vol-d0bc3f2c887f4072a9fda0b8915d94c1 - [tag:Name: "physical_volume_of_39c29f53-eac4-4f02-9781-90512cc7c563"] - would remove
demo10 - EC2Volume - vol-d1f066d8dac54ae59d087d7e9947e8a9 - [tag:Name: "Data volume for 4ccafc2e"] - would remove
demo10 - EC2Volume - vol-d9adb3f084cd4d588baa08690349b1f9 - [tag:Name: "volume_of_84854c9b-98aa-4f5b-926a-38b3398c3ad2"] - would remove
demo10 - EC2Volume - vol-db42e471b19f42b7835442545214bc1a - [tag:Name: "lb-tf-lb-20191126090616258000000002"] - would remove
demo10 - EC2Volume - vol-db80932fb47243efa67c9dd34223c647 - [tag:Name: "web-server-5"] - would remove
demo10 - EC2Volume - vol-dbea1d1083654d30a43366807a125aed - [tag:Name: "volume-555e9f8a"] - would remove

--- truncating long output ---
```