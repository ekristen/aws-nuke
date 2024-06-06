# Examples

## Basic usage

```bash
aws-nuke run --config config.yml
```

## Using a profile

!!! note
    This assumes you have configured your AWS credentials file with a profile named `my-profile`.

```bash
aws-nuke run --config config.yml --profile my-profile
```

## Using the prompt flags

!!! note
    This used be called the `--force` flag. It has been renamed to `--no-prompt` to better reflect its purpose.

!!! danger
    Running without prompts can be dangerous. Make sure you understand what you are doing before using these flags.

The following is an example of how you automate the command to run without any prompts of the user. This is useful
for running in a CI/CD pipeline.

```bash
aws-nuke run --config config.yml --no-prompt --prompt-delay 5
```