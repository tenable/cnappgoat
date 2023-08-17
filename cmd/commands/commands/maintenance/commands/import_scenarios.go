package commands

import (
	"fmt"
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var ImportScenariosCommand = &cli.Command{
	Name:  "import-scenarios",
	Usage: "maintenance subcommand to import scenarios from a local directory to the CNAPPgoat registry",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "directory",
			Usage:   "Directory to import scenarios from",
			Aliases: []string{"d"},
			Value:   "",
		},
	},
	Before: func(c *cli.Context) error {
		if err := common.CommandBefore(c); err != nil {
			return err
		}

		if c.String("directory") == "" {
			return fmt.Errorf("directory flag is required in import-scenarios subcommand")
		}

		return nil
	},
	Action: func(c *cli.Context) error {
		registry := c.Context.Value("CNAPPgoatModuleRegistry").(*cnappgoat.Registry)
		logrus.Infof("importing scenarios from %s", c.String("directory"))
		if _, err := registry.ImportScenarios(c.String("directory")); err != nil {
			return err
		}
		return nil
	},
}
