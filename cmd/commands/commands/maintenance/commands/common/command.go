package common

import (
	"context"
	
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/ermetic-research/CNAPPgoat/cmd/commands/common"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/urfave/cli/v2"
)

func CommandPulumiBefore(c *cli.Context) error {
	if err := common.CommandBefore(c); err != nil {
		return err
	}

	phdir, err := c.Context.Value("CNAPPgoatEngine").(*cnappgoat.Engine).Storage.GetPulumiHomeDir()
	if err != nil {
		return err
	}

	ph := auto.PulumiHome(phdir)
	if err != nil {
		return err
	}

	ws, err := auto.NewLocalWorkspace(c.Context, ph)
	if err != nil {
		return err
	}

	storage, err := cnappgoat.NewLocalStorage()
	if err != nil {
		return err
	}

	reg, err := cnappgoat.NewRegistry(storage)
	if err != nil {
		return err
	}

	c.Context = context.WithValue(c.Context, "CNAPPgoatWorkspace", ws)
	c.Context = context.WithValue(c.Context, "CNAPPgoatRegistry", reg)
	return nil
}
