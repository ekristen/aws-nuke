package list

import (
	"context"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/ekristen/aws-nuke/v3/pkg/commands/global"
	"github.com/ekristen/aws-nuke/v3/pkg/common"

	"github.com/ekristen/libnuke/pkg/registry"

	_ "github.com/ekristen/aws-nuke/v3/resources"
)

func execute(_ context.Context, c *cli.Command) error {
	var ls []string
	if c.Args().Len() > 0 {
		ls = registry.ExpandNames(c.Args().Slice())
	} else {
		ls = registry.GetNames()
	}

	slices.Sort(ls)

	for _, name := range ls {
		reg := registry.GetRegistration(name)

		if reg == nil {
			continue
		}

		if reg.AlternativeResource != "" {
			color.New(color.Bold).Printf("%-59s", name)
			color.New(color.FgCyan).Printf("native resource\n")
			color.New(color.Bold, color.FgYellow).Printf("  > %-55s", reg.AlternativeResource)
			color.New(color.FgHiBlue).Printf("alternative cloud-control resource\n")
		} else if strings.HasPrefix(reg.Name, "AWS::") {
			color.New(color.Bold).Printf("%-59s", name)
			color.New(color.FgHiMagenta).Printf("cloud-control resource\n")
		} else {
			color.New(color.Bold).Printf("%-59s", name)
			color.New(color.FgCyan).Printf("native resource\n")
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
