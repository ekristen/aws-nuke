package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/aws-nuke/v3/pkg/common"

	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/account"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/config"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/list"
	_ "github.com/ekristen/aws-nuke/v3/pkg/commands/nuke"

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

	app := cli.NewApp()
	app.Name = common.AppVersion.Name
	app.Usage = "remove everything from an aws account"
	app.Version = common.AppVersion.Summary
	app.Authors = []*cli.Author{
		{
			Name:  "Erik Kristensen",
			Email: "erik@erikkristensen.com",
		},
	}

	app.Commands = common.GetCommands()
	app.CommandNotFound = func(context *cli.Context, command string) {
		logrus.Fatalf("Command %s not found.", command)
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
