# CLI Shell Completion

The CLI supports shell completion for bash, and zsh. The completion script can be generated by running the 
following command:

```console
$ aws-nuke completion
```

By default, the shell is `bash` unless it can detect the shell you are using. You may specify the shell by using the 
`--shell` flag. 

The command will not install the completion script for you, but it will output the script to the console. You can
redirect the output to a file and source it in your shell to enable completion.

Command and flag completion is supported, however for flags that require a value, the completion will not provide a list
of possible values.

!!! warning
    For flag completion to work you often need to supply only the first `-` and press tab, depending on your shell
    configuration `--` followed by a tag will execute the command.

## Examples

!!! note
    The following are examples of commands you can run depending on your operating system and shell configuration.

### bash

```console
aws-nuke completion --shell bash > /etc/bash_completion.d/aws-nuke
```

### zsh

```console
aws-nuke completion --shell zsh > /usr/share/zsh/site-functions/_aws-nuke
```
