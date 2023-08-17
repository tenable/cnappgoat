package main

import (
	"context"
	"fmt"
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/commands/maintenance"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	// Populated by the build system
	// go build -ldflags "-X main.Version=1.0.0 -X main.BuildTime=$(date) -X main.GitCommit=$(git rev-parse HEAD)"
	version = "dev"
	date    = "N/A"
	commit  = "N/A"
)

func main() {
	app := &cli.App{
		Name:  "cnappgoat",
		Usage: "A multicloud open-source tool for deploying vulnerable-by-design cloud resources",
		Flags: common.MainFlags(),
		Before: func(c *cli.Context) error {
			debug := c.Bool("debug")
			common.SetDebug(debug)

			if c.Args().Len() == 0 {
				return nil
			}

			subcommandName := c.Args().First()
			if c.App.Command(subcommandName) == nil {
				_ = cli.ShowAppHelp(c)
				return fmt.Errorf("%s called with invalid subcommand: %s", c.App.Name, subcommandName)
			}

			moduleTable := common.GetDisplayTable()
			moduleTable.SetStyle(common.TableStyleCNAPPgoatMagenta)
			moduleTable.AppendHeader(table.Row{"#", "Scenario ID", "Scenario Name", "Platform", "Module", "Status"})
			moduleTable.AppendFooter(table.Row{""})

			storage, err := cnappgoat.NewLocalStorage()
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			reg, err := cnappgoat.NewRegistry(storage)
			if err != nil {
				return fmt.Errorf("failed to initialize registry: %w", err)
			}

			c.Context = context.WithValue(c.Context, "CNAPPgoatEngine", cnappgoat.NewEngine(reg, storage))
			c.Context = context.WithValue(c.Context, "CNAPPgoatModuleRegistry", reg)
			c.Context = context.WithValue(c.Context, "CNAPPgoatModuleStorage", storage)
			c.Context = context.WithValue(c.Context, "CNAPPgoatModuleTable", moduleTable)

			return nil
		},
		Version: fmt.Sprintf(
			"%s, date: %s, commit: %s\n",
			version,
			date,
			commit),
		Commands: []*cli.Command{
			commands.CleanCommand,
			commands.DescribeCommand,
			commands.DestroyCommand,
			commands.ListCommand,
			commands.OutputCommand,
			commands.ProvisionCommand,
			maintenance.MaintenanceCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
