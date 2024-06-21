package list

import (
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/common"

	_ "github.com/ekristen/aws-nuke/v3/resources"
	"github.com/ekristen/libnuke/pkg/registry"
)

func execute(c *cli.Context) error {
	ls := registry.GetNames()

	sort.Strings(ls)

	for _, name := range ls {
		if strings.HasPrefix(name, "AWS::") {
			continue
		}

		reg := registry.GetRegistration(name)

		if reg.AlternativeResource != "" {
			color.New(color.Bold).Printf("%-55s\n", name)
			color.New(color.Bold, color.FgYellow).Printf("  > %-55s", reg.AlternativeResource)
			color.New(color.FgCyan).Printf("alternative cloud-control resource\n")
		} else {
			color.New(color.Bold).Printf("%-55s\n", name)
		}
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
