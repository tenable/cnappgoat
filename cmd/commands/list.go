package commands

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

var ListCommand = &cli.Command{
	Name:   "list",
	Usage:  "List CNAPPgoat module scenarios",
	Flags:  common.CommandFlags(),
	Before: common.CommandUpdateBefore,
	Action: func(c *cli.Context) error {
		// NOTE: will panic if not initialized
		moduleTable := c.Context.Value("CNAPPgoatModuleTable").(table.Writer)
		reg := c.Context.Value("CNAPPgoatModuleRegistry").(*cnappgoat.Registry)

		scenarios, err := common.GetScenarios(
			c.Args(),
			reg,
			c.String("module"),
			c.String("platform"),
			c.String("state"))
		if err != nil {
			return err
		}

		for scenarioIndex, scenario := range scenarios {
			common.AppendScenarioToTable(scenario, moduleTable, scenarioIndex)
		}

		moduleTable.Render()

		return nil
	},
}
