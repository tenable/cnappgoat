package commands

import (
	"context"
	"fmt"

	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance/commands/common"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var ListSubCommand = &cli.Command{
	Name:   "list",
	Usage:  "maintenance subcommand to manage stacks",
	Before: common.CommandPulumiBefore,
	Action: func(c *cli.Context) error {
		r := c.Context.Value("CNAPPgoatRegistry").(*cnappgoat.Registry)
		e := c.Context.Value("CNAPPgoatEngine").(*cnappgoat.Engine)

		for _, scenario := range r.GetScenarios() {
			scenarioWorkDir := e.Storage.GetScenarioWorkingDir(scenario)

			ws, err := e.InitializeScenarioWorkspace(c.Context, scenario, scenarioWorkDir)
			if err != nil {
				logrus.WithError(err).Error("failed to initialize scenario workspace")
				continue
			}

			stacksNames, err := listStacks(c.Context, ws)
			if err != nil {
				return err
			}

			for _, s := range stacksNames {
				fmt.Println(s)
			}
		}

		return nil
	},
}

func listStacks(ctx context.Context, workspace auto.Workspace) ([]string, error) {
	stacks, err := workspace.ListStacks(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing stacks: %w", err)
	}

	var stackNames []string
	for _, stack := range stacks {
		stackNames = append(stackNames, stack.Name)
	}

	return stackNames, nil

}
