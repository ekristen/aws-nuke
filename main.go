package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"github.com/ekristen/aws-nuke/v3/pkg/common"

	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/account"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/completion"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/config"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/list"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/nuke"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/version"

	_ "github.com/ekristen/aws-nuke/v3/resources"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// log panics forces exit
			if _, ok := r.(*logrus.Entry); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	app := &cli.Command{
		Name:    common.AppVersion.Name,
		Usage:   "remove everything from an aws account",
		Version: common.AppVersion.Summary,
		Authors: []any{
			"Erik Kristensen <erik@erikkristensen.com>",
		},
		Commands: common.GetCommands(),
		CommandNotFound: func(ctx context.Context, command *cli.Command, s string) {
			logrus.Fatalf("Command %s not found.", s)
		},
		EnableShellCompletion: true,
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}
