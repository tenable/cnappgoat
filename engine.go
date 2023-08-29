package cnappgoat

import (
	"context"
	"errors"
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

type Engine struct {
	Storage  *LocalStorage
	Registry *Registry
}

type Option func(*ProvisionOptions)

type ProvisionOptions struct {
	AWSRegion   string
	AWSProfile  string
	AzureRegion string
	GCPRegion   string
}

func NewEngine(registry *Registry, storage *LocalStorage) *Engine {
	return &Engine{
		Storage:  storage,
		Registry: registry,
	}
}

func (e *Engine) CleanAll(ctx context.Context, force bool) error {
	cleanSuccessful := true
	for _, scenario := range e.Registry.ListScenarios() {
		if err := e.Clean(ctx, scenario, force); err != nil {
			logrus.WithError(err).Errorf("failed to Clean scenario %s", scenario.ScenarioParams.ID)
			cleanSuccessful = false
		}
	}
	if !cleanSuccessful {
		return errors.New("failed to Clean all scenarios")
	}
	return e.Storage.DeleteWorkingDir()
}

func (e *Engine) Destroy(
	ctx context.Context,
	scenario *Scenario,
	force bool) error {
	if scenario.State.State == NotDeployed || scenario.State.State == Destroyed {
		logrus.Infof("scenario %s is not deployed or already destroyed, skipping.", scenario.ScenarioParams.ID)
		return nil
	}

	stackName := getScenarioStackName(scenario)
	logrus.Infof("destroying stack %s", stackName)
	stack, err := e.initializeStackAndWorkspace(ctx, scenario, stackName)
	if err != nil {
		return e.writeErrorState(scenario, err, "failed to initialize stack and workspace")
	}

	if err :=
		e.refresh(
			ctx,
			stack,
			force,
			scenario); err != nil {
		return err
	}

	if _, err = stack.Destroy(ctx, optdestroy.ProgressStreams(logrus.StandardLogger().Out, logrus.NewEntry(logrus.StandardLogger()).WriterLevel(logrus.GetLevel()))); err != nil {
		return e.writeErrorState(scenario, err, "failed to destroy stack")
	}

	logrus.Info("successfully destroyed stack")
	if err = e.Registry.SetState(scenario, State{State: Destroyed}); err != nil {
		return fmt.Errorf("failed to write state to file: %w", err)
	}

	return nil
}

func (e *Engine) InitializeScenarioWorkspace(ctx context.Context, scenario *Scenario, scenarioWorkDir string) (auto.Workspace, error) {
	phDir, err := e.Storage.GetPulumiHomeDir()
	if err != nil {
		return nil, err
	}

	ph := auto.PulumiHome(phDir)
	backendURL, err := e.Storage.GetPulumiBackendURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get Pulumi backend URL: %w", err)
	}

	wd := auto.WorkDir(scenarioWorkDir)
	secretsProvider := auto.SecretsProvider("passphrase")

	envMap, err := environToMap()
	if err != nil {
		return nil, err
	}
	envMap["PULUMI_CONFIG_PASSPHRASE"] = "cnappgoat12345!"
	logrus.WithField("envMap", envMap).Debug("envMap")
	ws, err := auto.NewLocalWorkspace(ctx, wd, ph, secretsProvider, auto.EnvVars(envMap))
	if err != nil {
		return nil, fmt.Errorf("failed to create new local workspace: %w", err)
	}

	proj, err := workspace.LoadProject(e.Storage.GetProjectPath(scenario))
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	proj.Backend = &workspace.ProjectBackend{
		URL: backendURL,
	}
	if err = ws.SaveProjectSettings(ctx, proj); err != nil {
		return nil, fmt.Errorf("failed to save project settings: %w", err)
	}

	return ws, nil
}

func (e *Engine) Output(ctx context.Context, scenario *Scenario) (auto.OutputMap, error) {
	if scenario.State.State != Deployed && scenario.State.State != Error {
		logrus.Debugf("scenario %s is not deployed, skipping.", scenario.ScenarioParams.ID)
		return nil, nil
	}
	stackName := getScenarioStackName(scenario)
	stack, err := e.initializeStackAndWorkspace(ctx, scenario, stackName)
	if err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to initialize stack and workspace")
	}
	if err :=
		e.refresh(
			ctx,
			stack,
			false,
			scenario); err != nil {
		return nil, err
	}

	outputs, err := stack.Outputs(ctx)
	if err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to get outputs")
	}

	return outputs, nil
}

func (e *Engine) Provision(ctx context.Context, scenario *Scenario, force bool, opts ...Option) (auto.OutputMap, error) {
	options := &ProvisionOptions{}
	for _, opt := range opts {
		opt(options)
	}

	stackName := getScenarioStackName(scenario)
	scenarioWorkDir := e.Storage.GetScenarioWorkingDir(scenario)

	logrus.Infof("provisioning stack %s", stackName)
	ws, err := e.InitializeScenarioWorkspace(ctx, scenario, scenarioWorkDir)
	if err != nil {
		return nil, e.writeErrorState(scenario, err, fmt.Sprintf("failed to initialize workspace for stack %s", stackName))
	}

	stack, err := auto.UpsertStack(ctx, stackName, ws)
	if err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to initialize scenario stack")
	}

	if err = e.setStackConfigurationFromGlobalSettings(ctx, scenario, stack); err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to set stack configuration from global settings")
	}

	if err = e.setStackConfigurationFromProjectFile(ctx, scenario, stack); err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to set stack configuration from project file")
	}

	if err = e.setStackConfigurationFromOptions(ctx, scenario, stack, options); err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to set stack configuration from context")
	}

	if err :=
		e.refresh(
			ctx,
			stack,
			force,
			scenario); err != nil {
		return nil, err
	}

	res, err := stack.Up(ctx, optup.ProgressStreams(logrus.StandardLogger().Out, logrus.NewEntry(logrus.StandardLogger()).WriterLevel(logrus.GetLevel())))
	if err != nil {
		return nil, e.writeErrorState(scenario, err, fmt.Sprintf("failed to update stack %s", stackName))
	}

	logrus.Infof("successfully provision stack for scenario %s", scenario.ScenarioParams.ID)
	if err = e.Registry.SetState(scenario, State{State: Deployed}); err != nil {
		return nil, e.writeErrorState(scenario, err, "failed to write state to file")
	}

	return res.Outputs, nil
}

func (e *Engine) initializeStackAndWorkspace(ctx context.Context, scenario *Scenario, stackName string) (auto.Stack, error) {
	ws, err := e.InitializeScenarioWorkspace(ctx, scenario, e.Storage.GetScenarioWorkingDir(scenario))
	if err != nil {
		return auto.Stack{}, fmt.Errorf("failed to initialize scenario workspace: %w", err)
	}

	s, err := auto.UpsertStack(ctx, stackName, ws)
	if err != nil {
		return auto.Stack{}, e.writeErrorState(scenario, err, fmt.Sprintf("failed to initialize workspace for stack %s", stackName))
	}

	return s, nil
}

func (e *Engine) removeStack(ctx context.Context, scenario *Scenario) error {
	stackName := getScenarioStackName(scenario)

	w, err := e.InitializeScenarioWorkspace(ctx, scenario, e.Storage.GetScenarioWorkingDir(scenario))
	if err != nil {
		return e.writeErrorState(scenario, err, fmt.Sprintf("failed to initialized workspace for stack %s", stackName))
	}

	stacks, err := w.ListStacks(ctx)
	if err != nil {
		return e.writeErrorState(scenario, err, "failed to list stacks")
	}

	for _, stack := range stacks {
		if stack.Name == stackName {
			// We need this because you can have multiple stacks with the same name
			logrus.Infof("Removing stack %s", stackName)
			if err = w.RemoveStack(ctx, stackName); err != nil {
				return e.writeErrorState(scenario, err, "failed to remove stack")
			}

			logrus.Infof("successfully removed stack %s", stackName)
		}
	}

	if err = e.Registry.SetState(scenario, State{State: NotDeployed}); err != nil {
		return fmt.Errorf("failed to write state to file: %w", err)
	}

	return nil
}

func (e *Engine) setStackConfigurationFromProjectFile(ctx context.Context, scenario *Scenario, s auto.Stack) error {
	cnappGoatConfigParams, err := e.Storage.ReadCnappGoatConfig(scenario)
	if err != nil {
		return e.writeErrorState(scenario, err, "failed to read CNAPPgoat config")
	}

	for key, value := range cnappGoatConfigParams {
		if err = s.SetConfig(ctx, key, auto.ConfigValue{Value: value}); err != nil {
			return e.writeErrorState(scenario, err, "failed to set config")
		}
	}

	return nil
}

func (e *Engine) Clean(ctx context.Context, scenario *Scenario, force bool) error {
	if scenario.State.State != NotDeployed && scenario.State.State != Destroyed {
		if err := e.Destroy(ctx, scenario, force); err != nil {
			return fmt.Errorf("failed to destroy scenario %s: %v", scenario.ScenarioParams.ID, err)
		}
	}
	logrus.Info("removing stack")
	if err := e.removeStack(ctx, scenario); err != nil {
		return fmt.Errorf("failed to remove stack %s: %v", scenario.ScenarioParams.ID, err)
	}
	logrus.Infof("successfully cleaned scenario %s", scenario.ScenarioParams.ID)
	if err := e.Registry.SetState(scenario, State{State: NotDeployed}); err != nil {
		return err
	}
	return nil
}

func (e *Engine) setStackConfigurationFromOptions(ctx context.Context, scenario *Scenario, s auto.Stack, options *ProvisionOptions) error {
	if options.AWSRegion != "" {
		if err := s.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: options.AWSRegion}); err != nil {
			return e.writeErrorState(scenario, err, "failed to set config")
		}
	}

	if options.AzureRegion != "" {
		if err := s.SetConfig(ctx, "azure:location", auto.ConfigValue{Value: options.AzureRegion}); err != nil {
			return e.writeErrorState(scenario, err, "failed to set config")
		}
	}

	if options.GCPRegion != "" {
		if err := s.SetConfig(ctx, "gcp:region", auto.ConfigValue{Value: options.GCPRegion}); err != nil {
			return e.writeErrorState(scenario, err, "failed to set config")
		}
	}

	if options.AWSProfile != "" {
		if err := s.SetConfig(ctx, "aws:profile", auto.ConfigValue{Value: options.AWSProfile}); err != nil {
			return e.writeErrorState(scenario, err, "failed to set config")
		}
	}

	return nil
}

func (e *Engine) unlockStack(ctx context.Context, scenario *Scenario, s auto.Stack) error {
	logrus.Infof("stack %s is locked, unlocking", getScenarioStackName(scenario))
	if err := s.Cancel(ctx); err != nil {
		return e.writeErrorState(scenario, err, "failed to unlock stack")
	}

	logrus.Infof("successfully unlocked stack %s", getScenarioStackName(scenario))
	if _, err := s.Refresh(ctx); err != nil {
		return e.writeErrorState(scenario, err, "failed to refresh stack")
	}

	return nil
}

func (e *Engine) writeErrorState(scenario *Scenario, err error, msg string) error {
	if errState := e.Registry.SetState(scenario, State{State: Error, Msg: fmt.Sprintf("%v: %s", msg, err)}); errState != nil {
		return fmt.Errorf("failed to write state to file: %s: %w", msg, errors.Join(errState, err))
	}

	return fmt.Errorf("%v: %w", msg, err)
}

func (e *Engine) refresh(ctx context.Context, stack auto.Stack, force bool, scenario *Scenario) error {
	if _, err := stack.Refresh(ctx); err != nil {
		if auto.IsConcurrentUpdateError(err) {
			if !force {
				return e.writeErrorState(scenario, err, "failed to refresh stack."+
					" The stack is locked. Use the --force flag to unlock the stack. Note:"+
					" Note that this operation is very dangerous, and may leave the stack in an inconsistent state "+
					"if a resource operation was pending when the update was canceled. It might be preferable to run "+
					"the destroy command with the --force flag.")
			}

			if err := e.unlockStack(ctx, scenario, stack); err != nil {
				return err
			}
		} else {
			return e.writeErrorState(scenario, err, "failed to refresh stack")
		}
	}

	return nil
}

func getScenarioStackName(scenario *Scenario) string {
	return "cnappgoat_" + scenario.ScenarioParams.ID
}

func getValueFromEnvOrGcloudConfig(configKey string, vars ...string) string {
	if value := getEnv(vars...); value != "" {
		return value
	}
	return getGcloudValue(configKey)
}

func environToMap() (map[string]string, error) {
	envs := os.Environ()
	envMap := make(map[string]string, len(envs))

	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2) // Split only on the first '='
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid environment variable format: %s", env)
		}

		key := parts[0]
		value := parts[1]

		if key == "" {
			continue
			//return nil, fmt.Errorf("empty environment variable key found for value: %s", value)
		}

		envMap[key] = value
	}

	return envMap, nil
}
func (e *Engine) setStackConfigurationFromGlobalSettings(ctx context.Context, scenario *Scenario, s auto.Stack) error {
	if scenario.ScenarioParams.Platform == GCP {
		if _, err := exec.LookPath("gcloud"); err != nil {
			return e.writeErrorState(scenario, err, "Google Cloud CLI is not installed or configured correctly")
		}
		if project := getValueFromEnvOrGcloudConfig("core/project", "GOOGLE_PROJECT", "GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT"); project != "" {
			if err := s.SetConfig(ctx, "gcp:project", auto.ConfigValue{Value: project}); err != nil {
				return e.writeErrorState(scenario, err, "failed to set config")
			}
		}

		if region := getValueFromEnvOrGcloudConfig("compute/region", "GOOGLE_REGION", "GCLOUD_REGION", "CLOUDSDK_COMPUTE_REGION"); region != "" {
			if err := s.SetConfig(ctx, "gcp:region", auto.ConfigValue{Value: region}); err != nil {
				return e.writeErrorState(scenario, err, "failed to set config")
			}
		}

		if zone := getValueFromEnvOrGcloudConfig("compute/zone", "GOOGLE_ZONE", "GCLOUD_ZONE", "CLOUDSDK_COMPUTE_ZONE"); zone != "" {
			if err := s.SetConfig(ctx, "gcp:zone", auto.ConfigValue{Value: zone}); err != nil {
				return e.writeErrorState(scenario, err, "failed to set config")
			}
		}
	}

	if scenario.ScenarioParams.Platform == AWS {
		if _, err := exec.LookPath("aws"); err != nil {
			return e.writeErrorState(scenario, err, "AWS CLI is not installed or configured correctly")
		}
	}

	if scenario.ScenarioParams.Platform == Azure {
		if _, err := exec.LookPath("az"); err != nil {
			return e.writeErrorState(scenario, err, "Azure CLI is not installed or configured correctly")
		}
	}
	return nil
}

func getGcloudValue(configKey string) string {
	cmd := exec.Command("gcloud", "config", "get-value", configKey)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getEnv(vars ...string) string {
	for _, v := range vars {
		if value := os.Getenv(v); value != "" {
			return value
		}
	}
	return ""
}
