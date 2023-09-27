package commands

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var CleanCommand = &cli.Command{
	Name:   "clean",
	Usage:  "clean and remove all created resources and delete all scenarios and any related files",
	Flags:  common.CommandFlags(),
	Before: common.CommandUpdateBefore,
	Action: func(c *cli.Context) error {
		moduleTable := c.Context.Value("CNAPPgoatModuleTable").(table.Writer)
		engine := c.Context.Value("CNAPPgoatEngine").(*cnappgoat.Engine)
		reg := c.Context.Value("CNAPPgoatModuleRegistry").(*cnappgoat.Registry)
		if c.Args().Len() == 0 && c.String("module") == "" && c.String("platform") == "" && c.String("state") == "" {
			if err := engine.CleanAll(c.Context, c.Bool("force")); err != nil {
				return err
			}
			logrus.Infof("successfully cleaned all scenarios, and removed all created resources")
			logrus.Infof("goodbye!")
			return nil
		}
		scenarios, err :=
			common.GetScenarios(
				c.Args(),
				reg,
				c.String("module"),
				c.String("platform"),
				c.String("state"))
		if err != nil {
			return err
		}
		if len(scenarios) >= 3 {
			ok, err := common.ConfirmForAllScenarios("clean", len(scenarios))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		for scenarioIndex, scenario := range scenarios {
			logrus.Infof("destroying scenario: %s", scenario.Name)
			if err := engine.Clean(c.Context, scenario, c.Bool("force")); err != nil {
				logrus.WithError(err).Error("failed to clean scenario")
			}

			common.AppendScenarioToTable(scenario, moduleTable, scenarioIndex)
		}

		moduleTable.Render()
		return nil
	},
}
