package version

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/common"
)

func execute(c *cli.Context) error {
	fmt.Println(common.AppVersion.Name, "version", common.AppVersion.Summary)

	return nil
}

func init() {
	cmd := &cli.Command{
		Name:        "version",
		Usage:       "displays the version",
		Description: "displays the version of aws-nuke",
		Flags:       global.Flags(),
		Before:      global.Before,
		Action:      execute,
	}

	common.RegisterCommand(cmd)
}
