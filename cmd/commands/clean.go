package commands

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var CleanCommand = &cli.Command{
	Name:   "clean",
	Usage:  "clean and remove all created resources and delete all scenarios and any related files",
	Flags:  common.CommandFlags(),
	Before: common.CommandBefore,
	Action: func(c *cli.Context) error {
		engine := c.Context.Value("CNAPPgoatEngine").(*cnappgoat.Engine)

		if err := engine.CleanAll(c.Context, c.Bool("force")); err != nil {
			return err
		}

		logrus.Infof("successfully cleaned all scenarios, and removed all created resources")
		logrus.Infof("goodbye!")
		return nil
	},
}
