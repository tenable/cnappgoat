package plugins

import (
	"errors"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands/plugins/commands"
	"github.com/urfave/cli/v2"
)

var PluginsCommand = &cli.Command{
	Name:  "plugins",
	Usage: "maintenance subcommand to manage plugins",
	Action: func(c *cli.Context) error {
		return errors.New("plugins command must be followed by a subcommand")
	},
	Subcommands: []*cli.Command{
		commands.ListCommand,
	},
}
