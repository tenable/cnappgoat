package commands

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var DestroyCommand = &cli.Command{
	Name:   "destroy",
	Usage:  "Destroy CNAPPgoat module scenarios",
	Flags:  common.CommandFlags(),
	Before: common.CommandBefore,
	Action: func(c *cli.Context) error {
		moduleTable := c.Context.Value("CNAPPgoatModuleTable").(table.Writer)
		engine := c.Context.Value("CNAPPgoatEngine").(*cnappgoat.Engine)
		reg := c.Context.Value("CNAPPgoatModuleRegistry").(*cnappgoat.Registry)

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
			ok, err := common.ConfirmForAllScenarios("destroy", len(scenarios))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		for scenarioIndex, scenario := range scenarios {
			logrus.Infof("destroying scenario: %s", scenario.Name)
			if err := engine.Destroy(c.Context, scenario, c.Bool("force")); err != nil {
				logrus.WithError(err).Error("failed to destroy scenario")
			}

			common.AppendScenarioToTable(scenario, moduleTable, scenarioIndex)
		}

		moduleTable.Render()
		return nil
	},
}
