package commands

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/glamour"
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/urfave/cli/v2"
	"text/template"
)

const scenarioTemplate = `# {{.FriendlyName}}

## Description	
{{.Desc}}

### Attributes

**ID**: {{.Id}}

**Platform**: {{.Platform}}

**Module**: {{.Module}}

**runtime**: {{.runtime}}

**State**: {{.State}}
 
**Location**: {{.SrcDir}}

***
`

var DescribeCommand = &cli.Command{
	Name:   "describe",
	Usage:  "describe a scenario",
	Flags:  common.CommandFlags(),
	Before: common.CommandUpdateBefore,
	Action: func(c *cli.Context) error {
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

		for _, scenario := range scenarios {
			data :=
				map[string]interface{}{
					"Desc":         scenario.ScenarioParams.Description,
					"FriendlyName": scenario.ScenarioParams.FriendlyName,
					"Id":           scenario.ScenarioParams.ID,
					"Module":       scenario.ScenarioParams.Module,
					"Platform":     scenario.ScenarioParams.Platform,
					"runtime":      scenario.Runtime.Name,
					"SrcDir":       scenario.SrcDir,
					"State":        scenario.State.State,
				}

			out, err := glamour.Render(getScenarioTemplate(data), "auto")
			if err != nil {
				return err
			}

			fmt.Print(out)
		}

		return nil
	},
}

func getScenarioTemplate(m map[string]interface{}) string {
	buf := &bytes.Buffer{}
	if err :=
		template.Must(template.New("markdown").
			Parse(scenarioTemplate)).
			Execute(buf, m); err != nil {
		panic(err)
	}

	return buf.String()
}
