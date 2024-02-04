package list

import (
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	"github.com/ekristen/aws-nuke/pkg/commands/global"
	"github.com/ekristen/aws-nuke/pkg/common"

	"github.com/ekristen/libnuke/pkg/resource"

	_ "github.com/ekristen/aws-nuke/resources"
)

func execute(c *cli.Context) error {
	ls := resource.GetNames()

	sort.Strings(ls)

	for _, name := range ls {
		if strings.HasPrefix(name, "AWS::") {
			continue
		}

		reg := resource.GetRegistration(name)

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
