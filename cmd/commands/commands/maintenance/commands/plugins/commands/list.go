package commands

import (
	"fmt"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands/common"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/urfave/cli/v2"
)

var ListCommand = &cli.Command{
	Name:   "list",
	Usage:  "list plugins",
	Before: common.CommandPulumiBefore,
	Action: func(c *cli.Context) error {
		w := c.Context.Value("CNAPPgoatWorkspace").(*auto.LocalWorkspace)

		plugins, err := w.ListPlugins(c.Context)
		if err != nil {
			return fmt.Errorf("failed listing plugins: %w", err)
		}

		for _, plugin := range plugins {
			fmt.Println(plugin.Spec().String())
		}

		return nil
	},
}
