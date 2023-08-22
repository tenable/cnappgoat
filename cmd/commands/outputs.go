package commands

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"sort"
)

var OutputCommand = &cli.Command{
	Name:   "output",
	Usage:  "print scenario outputs",
	Flags:  common.OutputFlags(),
	Before: common.CommandUpdateBefore,
	Action: func(c *cli.Context) error {
		outputTable := common.GetDisplayTable()
		outputTable.SetStyle(common.TableStyleCNAPPgoatMagenta)
		outputTable.AppendHeader(table.Row{"#", "Scenario ID", "Key", "Value"})
		outputTable.AppendFooter(table.Row{""})
		tableIndex := 0
		engine := c.Context.Value("CNAPPgoatEngine").(*cnappgoat.Engine)
		reg := c.Context.Value("CNAPPgoatModuleRegistry").(*cnappgoat.Registry)

		scenarios, err :=
			common.GetScenarios(
				c.Args(),
				reg,
				c.String("module"),
				c.String("platform"),
				c.String("status"))
		if err != nil {
			return err
		}
		for _, scenario := range scenarios {
			if outputs, err := engine.Output(c.Context, scenario); err != nil {
				logrus.WithError(err).Error("failed to retrieve scenario outputs")
			} else {
				if err != nil {
					logrus.WithError(err).Error("failed to marshal outputs")
				} else if len(outputs) > 0 {
					tableIndex++
					firstOutput := true
					keys := make([]string, 0, len(outputs))
					for key := range outputs {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					for _, key := range keys {
						value := outputs[key]
						if value.Secret && !c.Bool("show-secrets") {
							value.Value = "********"
						}
						if firstOutput {
							outputTable.AppendRow(table.Row{tableIndex, scenario.ScenarioParams.ID, key, value.Value})
							firstOutput = false
						} else {
							outputTable.AppendRow(table.Row{"", "", key, value.Value})
						}
					}
				}
			}
		}
		outputTable.Render()
		return nil
	},
}
