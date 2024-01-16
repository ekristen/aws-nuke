# First Run

## First Configuration

First you need to create a config file for *aws-nuke*. This is a minimal one:

```yaml
regions:
  - global
  - us-east-1

account-blocklist:
- "999999999999" # production

accounts:
  "000000000000": {} # aws-nuke-example
```

## First Run (Dry Run)

With this config we can run *aws-nuke*:

```bash
$ aws-nuke nuke -c config/nuke-config.yaml
aws-nuke version v1.0.39.gc2f318f - Fri Jul 28 16:26:41 CEST 2017 - c2f318f37b7d2dec0e646da3d4d05ab5296d5bce

Do you really want to nuke the account with the ID 000000000000 and the alias 'aws-nuke-example'?
Do you want to continue? Enter account alias to continue.
> aws-nuke-example

us-east-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - would remove
us-east-1 - EC2Instance - 'i-01b489457a60298dd' - would remove
us-east-1 - EC2KeyPair - 'test' - would remove
us-east-1 - EC2NetworkACL - 'acl-6482a303' - cannot delete default VPC
us-east-1 - EC2RouteTable - 'rtb-ffe91e99' - would remove
us-east-1 - EC2SecurityGroup - 'sg-220e945a' - cannot delete group 'default'
us-east-1 - EC2SecurityGroup - 'sg-f20f958a' - would remove
us-east-1 - EC2Subnet - 'subnet-154d844e' - would remove
us-east-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - would remove
us-east-1 - EC2VPC - 'vpc-c6159fa1' - would remove
us-east-1 - IAMUserAccessKey - 'my-user -> ABCDEFGHIJKLMNOPQRST' - would remove
us-east-1 - IAMUserPolicyAttachment - 'my-user -> AdministratorAccess' - [UserName: "my-user", PolicyArn: "arn:aws:iam::aws:policy/AdministratorAccess", PolicyName: "AdministratorAccess"] - would remove
us-east-1 - IAMUser - 'my-user' - would remove
Scan complete: 13 total, 11 nukeable, 2 filtered.

Would delete these resources. Provide --no-dry-run to actually destroy resources.
```

As we see, *aws-nuke* only lists all found resources and exits. This is because the `--no-dry-run` flag is missing.
Also, it wants to delete the administrator. We don't want to do this, because we use this user to access our account.
Therefore, we have to extend the config, so it ignores this user:

```yaml
regions:
- us-east-1

account-blocklist:
- "999999999999" # production

accounts:
  "000000000000": # aws-nuke-example
    filters:
      IAMUser:
        - "my-user"
      IAMUserPolicyAttachment:
        - "my-user -> AdministratorAccess"
      IAMUserAccessKey:
        - "my-user -> ABCDEFGHIJKLMNOPQRST"
```

## Second Run (No Dry Run)

!!! warning
This will officially remove resources from your AWS account. Make sure you really want to do this!

```bash
$ aws-nuke nuke -c config/nuke-config.yml --no-dry-run
aws-nuke version v1.0.39.gc2f318f - Fri Jul 28 16:26:41 CEST 2017 - c2f318f37b7d2dec0e646da3d4d05ab5296d5bce

Do you really want to nuke the account with the ID 000000000000 and the alias 'aws-nuke-example'?
Do you want to continue? Enter account alias to continue.
> aws-nuke-example

us-east-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - would remove
us-east-1 - EC2Instance - 'i-01b489457a60298dd' - would remove
us-east-1 - EC2KeyPair - 'test' - would remove
us-east-1 - EC2NetworkACL - 'acl-6482a303' - cannot delete default VPC
us-east-1 - EC2RouteTable - 'rtb-ffe91e99' - would remove
us-east-1 - EC2SecurityGroup - 'sg-220e945a' - cannot delete group 'default'
us-east-1 - EC2SecurityGroup - 'sg-f20f958a' - would remove
us-east-1 - EC2Subnet - 'subnet-154d844e' - would remove
us-east-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - would remove
us-east-1 - EC2VPC - 'vpc-c6159fa1' - would remove
us-east-1 - IAMUserAccessKey - 'my-user -> ABCDEFGHIJKLMNOPQRST' - filtered by config
us-east-1 - IAMUserPolicyAttachment - 'my-user -> AdministratorAccess' - [UserName: "my-user", PolicyArn: "arn:aws:iam::aws:policy/AdministratorAccess", PolicyName: "AdministratorAccess"] - would remove
us-east-1 - IAMUser - 'my-user' - filtered by config
Scan complete: 13 total, 8 nukeable, 5 filtered.

Do you really want to nuke these resources on the account with the ID 000000000000 and the alias 'aws-nuke-example'?
Do you want to continue? Enter account alias to continue.
> aws-nuke-example

us-east-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - failed
us-east-1 - EC2Instance - 'i-01b489457a60298dd' - triggered remove
us-east-1 - EC2KeyPair - 'test' - triggered remove
us-east-1 - EC2RouteTable - 'rtb-ffe91e99' - failed
us-east-1 - EC2SecurityGroup - 'sg-f20f958a' - failed
us-east-1 - EC2Subnet - 'subnet-154d844e' - failed
us-east-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - failed
us-east-1 - EC2VPC - 'vpc-c6159fa1' - failed
us-east-1 - S3Object - 's3://rebuy-terraform-state-138758637120/run-terraform.lock' - triggered remove

Removal requested: 2 waiting, 6 failed, 5 skipped, 0 finished

us-east-1 - EC2DHCPOption - 'dopt-bf2ec3d8' - failed
us-east-1 - EC2Instance - 'i-01b489457a60298dd' - waiting
us-east-1 - EC2KeyPair - 'test' - removed
us-east-1 - EC2RouteTable - 'rtb-ffe91e99' - failed
us-east-1 - EC2SecurityGroup - 'sg-f20f958a' - failed
us-east-1 - EC2Subnet - 'subnet-154d844e' - failed
us-east-1 - EC2Volume - 'vol-0ddfb15461a00c3e2' - failed
us-east-1 - EC2VPC - 'vpc-c6159fa1' - failed

Removal requested: 1 waiting, 6 failed, 5 skipped, 1 finished

--- truncating long output ---
```

As you see *aws-nuke* now tries to delete all resources which aren't filtered, without caring about the dependencies
between them. This results in API errors which can be ignored. These errors are shown at the end of the *aws-nuke* run,
if they keep to appear.

*aws-nuke* retries deleting all resources until all specified ones are deleted
or until there are only resources with errors left.

