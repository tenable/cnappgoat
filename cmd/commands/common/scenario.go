package common

import (
	"fmt"
	cnappgoat "github.com/ermetic-research/CNAPPgoat"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"strings"
)

func GetScenarios(
	args cli.Args,
	reg *cnappgoat.Registry,
	module string,
	platform string) ([]*cnappgoat.Scenario, error) {
	if scenarios := getScenarios(
		args,
		reg,
		module,
		platform); len(scenarios) > 0 {
		return scenarios, nil
	}

	return nil, fmt.Errorf("no scenarios found")
}

func ConfirmForAllScenarios(action string, scenariosCount int) (bool, error) {
	fmt.Printf("You are about to %s %d scenarios\n", action, scenariosCount)
	fmt.Println("This may take a while, and may incur charges on your cloud provider account")
	fmt.Println("In addition, this may work partially, or not at all, depending on your cloud provider and region")
	fmt.Println("Are you sure you want to proceed?")
	fmt.Println("Enter yes/no: ")

	for {
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return false, fmt.Errorf("failed to get user input: %w", err)
		}

		switch strings.ToLower(response) {
		case "yes":
			logrus.Infof("proceeding with %s action", action)
			return true, nil
		case "no":
			logrus.Infof("aborting %s", action)
			return false, nil
		default:
			logrus.Infof("invalid response: %s. Please enter 'yes' or 'no'.", response)
		}
	}
}

func getScenarios(
	args cli.Args,
	reg *cnappgoat.Registry,
	module string,
	platform string) (scenarios []*cnappgoat.Scenario) {
	if args.Len() > 0 {
		for _, arg := range separateArgs(args.Slice()) {
			scenario, ok := reg.GetScenario(arg)
			if !ok {
				logrus.Errorf("failed to find scenario: %s", arg)
				continue
			}

			if module != "" &&
				string(scenario.ScenarioParams.Module) != strings.ToUpper(module) {
				logrus.Errorf("Skipping scenario %s because it is not part of module %s", scenario.ScenarioParams.ID, module)
				continue
			}

			if platform != "" &&
				string(scenario.ScenarioParams.Platform) != strings.ToUpper(platform) {
				logrus.Errorf("Skipping scenario %s because it is not part of platform %s", scenario.ScenarioParams.ID, platform)
				continue
			}

			scenarios = append(scenarios, scenario)
		}

		return
	}

	if module != "" && platform != "" {
		logrus.Debugf("list all scenarios for %s/%s", module, platform)
		scenarios = reg.ListScenariosByModuleAndPlatform(cnappgoat.Module(module), cnappgoat.Platform(platform))
		return
	}

	if platform != "" {
		logrus.Debugf("list all scenarios for %s", platform)
		scenarios = reg.ListScenariosByPlatform(cnappgoat.Platform(platform))
		return
	}

	if module != "" {
		logrus.Debugf("list all scenarios for %s", module)
		scenarios = reg.ListScenariosByModule(cnappgoat.Module(module))
		return
	}

	logrus.Debugf("list all scenarios for all modules")
	scenarios = reg.ListScenarios()
	return
}

func separateArgs(args []string) (positional []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		} else {
			positional = append(positional, arg)
		}
	}

	return
}
