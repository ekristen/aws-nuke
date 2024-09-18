package completion

import (
	"embed"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/common"
)

//go:embed files/*
var files embed.FS

func execute(c *cli.Context) error {
	var autocomplete []byte
	var err error
	switch c.String("shell") {
	case "bash":
		autocomplete, err = files.ReadFile("files/bash_autocomplete")
	case "zsh":
		autocomplete, err = files.ReadFile("files/zsh_autocomplete")
	}

	if err != nil {
		return err
	}

	fmt.Println(string(autocomplete))

	return nil
}

func init() {
	shellValue := "bash"
	shellActual := os.Getenv("SHELL")
	if strings.Contains(shellActual, "zsh") {
		shellValue = "zsh"
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "shell",
			Usage: "shell to generate completion script for",
			Value: shellValue,
			Action: func(c *cli.Context, val string) error {
				validShells := []string{"bash", "zsh"}
				if !slices.Contains(validShells, val) {
					return fmt.Errorf("unsupported shell %s", val)
				}

				return nil
			},
		},
	}

	cmd := &cli.Command{
		Name:        "completion",
		Usage:       "generate shell completion script",
		Description: "generate shell completion script",
		Flags:       append(flags, global.Flags()...),
		Before:      global.Before,
		Action:      execute,
	}

	common.RegisterCommand(cmd)
}
