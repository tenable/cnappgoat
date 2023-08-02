package stacks

import (
	"errors"

	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands/stacks/commands"
	"github.com/urfave/cli/v2"
)

var StacksCommand = &cli.Command{
	Name:  "stacks",
	Usage: "maintenance subcommand to manage stacks",
	Action: func(c *cli.Context) error {
		return errors.New("stacks command must be followed by a subcommand")
	},
	Subcommands: []*cli.Command{
		commands.ListSubCommand,
	},
}
