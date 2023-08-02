package commands

import (
	"fmt"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/ermetic-research/CNAPPgoat/infra"
	"github.com/urfave/cli/v2"
)

var GetScenariosCommand = &cli.Command{
	Name:  "get-scenarios",
	Usage: "maintenance subcommand to download scenarios",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "directory",
			Usage:   "Directory to download scenarios to",
			Aliases: []string{"d"},
			Value:   "",
		},
	},
	Before: func(c *cli.Context) error {
		if err := common.CommandBefore(c); err != nil {
			return err
		}

		if c.String("directory") == "" {
			return fmt.Errorf("directory flag is required in downloadScenarios subcommand")
		}

		return nil
	},
	Action: func(c *cli.Context) error {
		if _, err := infra.GitDownloadScenarios(c.String("directory")); err != nil {
			return err
		}

		return nil
	},
}
