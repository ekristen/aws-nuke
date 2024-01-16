package list

import (
	"fmt"
	"github.com/ekristen/aws-nuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/aws-nuke/pkg/commands/global"
	"github.com/ekristen/aws-nuke/pkg/common"

	_ "github.com/ekristen/aws-nuke/resources"
)

func execute(c *cli.Context) error {
	ls := resource.GetListersForScope(nuke.Account)

	for name, _ := range ls {
		fmt.Println(name)
	}

	return nil
}

func init() {
	cmd := &cli.Command{
		Name:    "resource-types",
		Aliases: []string{"list-resources"},
		Usage:   "list available resources to nuke",
		Flags:   global.Flags(),
		Before:  global.Before,
		Action:  execute,
	}

	common.RegisterCommand(cmd)
}
