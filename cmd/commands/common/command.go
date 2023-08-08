package common

import (
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/urfave/cli/v2"
)

func CommandBefore(c *cli.Context) error {
	SetDebug(c.Bool("debug"))

	if module := c.String("module"); module != "" {
		if _, err := cnappgoat.ModuleFromString(module); err != nil {
			return err
		}
	}
	if platform := c.String("platform"); platform != "" {
		if _, err := cnappgoat.PlatformFromString(platform); err != nil {
			return err
		}
	}
	if state := c.String("state"); state != "" {
		if _, err := cnappgoat.StateFromString(state); err != nil {
			return err
		}
	}

	return nil
}

func CommandUpdateBefore(c *cli.Context) error {
	if err := CommandBefore(c); err != nil {
		return err
	}

	reg := c.Context.Value("CNAPPgoatModuleRegistry").(*cnappgoat.Registry)

	if err := reg.UpdateRegistryFromGit(); err != nil {
		return err
	}

	return nil
}

func MainFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode",
			Value: false,
		},
	}
}

func CommandFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode",
			Value: false,
		},
		&cli.BoolFlag{
			Name:    "force",
			Usage:   "Enable force mode (unlock locked stacks with pulumi cancel)",
			Aliases: []string{"f"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "module",
			Usage:   "CNAPPgoat module to operate on, e.g. 'CIEM, 'CSPM', 'CWPP', 'IAC'",
			Aliases: []string{"m"},
		},
		&cli.StringFlag{
			Name:    "platform",
			Usage:   "CNAPPgoat platform to operate on, e.g. 'AWS, 'AZURE', 'GCP'",
			Aliases: []string{"p"},
		},
		&cli.StringFlag{
			Name:    "state",
			Usage:   "CNAPPgoat state to filter on, e.g. 'Deployed', 'Destroyed', 'Error', 'Not Deployed'",
			Aliases: []string{"s"},
		},
	}
}
