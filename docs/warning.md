# Caution and Warning

!!! danger
    Be aware that *aws-nuke* is a very destructive tool, hence you have to be very careful while using it. Otherwise,
    you might delete production data.

!!! warning
    **We strongly advise you to not run this application on any AWS account, where you cannot afford to lose
    all resources.**

To reduce the blast radius of accidents, there are some safety precautions:

1. By default, *aws-nuke* only lists all nuke-able resources. You need to add `--no-dry-run` to actually delete
   resources.
2. *aws-nuke* asks you twice to confirm the deletion by entering the account alias. The first time is directly
   after the start and the second time after listing all nuke-able resources.
3. To avoid just displaying a account ID, which might gladly be ignored by humans, it is required to actually set
   an [Account Alias](https://docs.aws.amazon.com/IAM/latest/UserGuide/console_account-alias.html) for your account. Otherwise, *aws-nuke* will abort.
4. The Account Alias must not contain the string `prod`. This string is hardcoded, and it is recommended to add it
   to every actual production account (e.g. `mycompany-production-ecr`).
5. The config file contains a blocklist field. If the Account ID of the account you want to nuke is part of this
   blocklist, *aws-nuke* will abort. It is recommended, that you add every production account to this blocklist.
6. To ensure you don't just ignore the blocklisting feature, the blocklist must contain at least one Account ID.
7. The config file contains account specific settings (e.g. filters). The account you want to nuke must be explicitly
   listed there.
8. To ensure to not accidentally delete a random account, it is required to specify a config file. It is recommended
   to have only a single config file and add it to a central repository. This way the account blocklist is way
   easier to manage and keep up to date.

Feel free to create an issue, if you have any ideas to improve the safety procedures.

