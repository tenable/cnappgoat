package maintenance

import (
	"errors"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands/plugins"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands/stacks"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/urfave/cli/v2"
)

var MaintenanceCommand = &cli.Command{
	Name:  "maintenance",
	Usage: "CNAPPgoat maintenance command to perform maintenance tasks",
	Flags: common.CommandFlags(),
	Action: func(c *cli.Context) error {
		return errors.New("maintenance command must be followed by a subcommand")
	},
	Subcommands: []*cli.Command{
		plugins.PluginsCommand,
		stacks.StacksCommand,
		commands.GetScenariosCommand,
		commands.ImportScenariosCommand,
	},
	Hidden: true,
}
