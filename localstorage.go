package cnappgoat

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ermetic-research/CNAPPgoat/infra"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const stateFile = "state.yaml"
const gitMetadataFile = ".gitMetadataFile"

type gitMetadata struct {
	CommitHash string `yaml:"commitHash"`
	Date       string `yaml:"date"`
}

type LocalStorage struct {
	WorkingDir string
}

func NewLocalStorage() (*LocalStorage, error) {
	workDir, err := getLocalWorkDirPath()
	if err != nil {
		return nil, fmt.Errorf("unable to get local working directory path: %w", err)
	}

	return &LocalStorage{
		WorkingDir: workDir,
	}, nil
}

func (l *LocalStorage) DeleteWorkingDir() error {
	return os.RemoveAll(l.WorkingDir)
}

func (l *LocalStorage) GetProjectPath(scenario *Scenario) string {
	return filepath.Join(l.GetScenarioWorkingDir(scenario), "Pulumi.yaml")
}

func (l *LocalStorage) GetPulumiBackendURL() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user home directory: %w", err)
	}

	pulumiBackendPath := filepath.ToSlash(filepath.Join(homeDir, ".cnappgoat"))

	u := &url.URL{
		Scheme: "file",
		Path:   pulumiBackendPath,
	}

	return u.String(), nil
}

func (l *LocalStorage) GetPulumiHomeDir() (string, error) {
	localWorkDir, err := getLocalWorkDirPath()
	if err != nil {
		return "", fmt.Errorf("unable to get local working directory:  %w", err)
	}

	return filepath.Join(localWorkDir, ".pulumi"), nil
}

func (l *LocalStorage) GetScenarioWorkingDir(scenario *Scenario) string {
	return filepath.Join(
		l.WorkingDir,
		"scenarios",
		strings.ToLower(string(scenario.ScenarioParams.Module)),
		strings.ToLower(string(scenario.ScenarioParams.Platform)),
		strings.ToLower(strings.Join(strings.Split(scenario.ScenarioParams.ID, "-")[2:], "-")),
	)
}

func (l *LocalStorage) UpdateScenariosFromGit() (map[string]*Scenario, error) {
	metadata, err := l.getCurrentGitMetadata()
	today := time.Now().Format("2006-01-02")
	if metadata != nil && metadata.Date == today {
		logrus.Debug("scenarios downloaded today, skipping git update")
		scenarios, err := l.loadScenariosFromWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("unable to load scenarios from working directory:  %w", err)
		}
		return scenarios, nil
	}

	tempDir, remoteHash, err := infra.GitDownloadScenariosToTempDir()
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			logrus.WithError(err).Error("unable to remove temporary directory")
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("unable to clone git repository:  %w", err)
	}

	if metadata != nil && remoteHash == metadata.CommitHash {
		logrus.Debug("remote hash matches local hash, skipping git update")
		scenarios, err := l.loadScenariosFromWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("unable to load scenarios from working directory:  %w", err)
		}
		return scenarios, nil
	}

	scenarios, err := l.updateScenariosFromFolder(tempDir)
	if err != nil {
		return nil, fmt.Errorf("unable to update scenario folder:  %w", err)
	}

	if err := l.writeGitMetadata(remoteHash, today); err != nil {
		return nil, fmt.Errorf("unable to save git metadata:  %w", err)
	}
	return scenarios, nil
}

func (l *LocalStorage) LoadScenariosFromWorkingDir() (map[string]*Scenario, error) {
	return l.loadScenarios(l.WorkingDir)
}

func (l *LocalStorage) ReadCnappGoatConfig(scenario *Scenario) (map[string]string, error) {
	scenarioWorkDir := l.GetScenarioWorkingDir(scenario)
	path := filepath.Join(scenarioWorkDir, "Pulumi.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Pulumi.yaml: %w", err)
	}

	var cnappGoatConfig Scenario
	if err = yaml.Unmarshal(data, &cnappGoatConfig); err != nil {
		return nil, fmt.Errorf("failed to parse Pulumi.yaml: %w", err)
	}

	return cnappGoatConfig.ScenarioParams.Config, nil
}

func (l *LocalStorage) WriteStateToFile(scenario *Scenario, state State) error {
	// Create the file path.
	filePath := filepath.Join(l.GetScenarioWorkingDir(scenario), stateFile)
	// Marshal the struct to YAML.
	data, err := yaml.Marshal(&state)
	if err != nil {
		return fmt.Errorf("failed to marshal state to YAML:  %w", err)
	}

	// Create the file if it doesn't exist.
	if !l.fileExists(filePath) {
		// Create a new file.
		_, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create state file: %w", err)
		}
	}
	// Write the YAML data to the file.
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state to file: %w", err)
	}

	return nil
}

func (l *LocalStorage) updateScenariosFromFolder(scenariosFullPath string) (map[string]*Scenario, error) {
	if !l.fileExists(scenariosFullPath) {
		return nil, fmt.Errorf("scenario folder does not exist, cannot perform UpdateScenarioFolder: %s", scenariosFullPath)
	}
	scenariosFromScenarioDir, err := l.loadScenarios(scenariosFullPath)
	var scenariosFromWorkDir map[string]*Scenario
	if l.WorkingDirectoryExists() {
		scenariosFromWorkDir, err = l.loadScenarios(l.WorkingDir)
		if err != nil {
			return nil, fmt.Errorf("unable to load scenarios from working directory:  %w", err)
		}
	} else {
		scenariosFromWorkDir = make(map[string]*Scenario)
		err = os.MkdirAll(l.WorkingDir, 0755)
	}

	for _, scenarioFromScenariosDir := range scenariosFromScenarioDir {
		exists := false
		for _, scenarioFromWorkDir := range scenariosFromWorkDir {
			if scenarioFromWorkDir.ScenarioParams.ID == scenarioFromScenariosDir.ScenarioParams.ID {
				exists = true
				if scenarioFromWorkDir.Hash != scenarioFromScenariosDir.Hash {
					err = l.copyScenario(scenarioFromScenariosDir)
					if err != nil {
						return nil, fmt.Errorf("unable to copy scenario to working directory:  %w", err)
					}
					scenariosFromWorkDir[scenarioFromScenariosDir.ScenarioParams.ID] = scenarioFromScenariosDir
				} else {
					logrus.Debugf("Scenario %v exists in the working directory and has not changed. Skipping.", scenarioFromWorkDir.ScenarioParams.ID)
				}
			}
		}
		if !exists {
			// if the scenario does not exist in the working directory, copy it over
			logrus.Debugf("Scenario %v does not exist in the working directory. Copying over.", scenarioFromScenariosDir.ScenarioParams.ID)
			err = l.copyScenario(scenarioFromScenariosDir)
			if err != nil {
				return nil, fmt.Errorf("unable to copy scenario to working directory:  %w", err)
			}
			scenariosFromWorkDir[scenarioFromScenariosDir.ScenarioParams.ID] = scenarioFromScenariosDir
		}
	}
	return scenariosFromWorkDir, err
}

func (l *LocalStorage) WorkingDirectoryExists() bool {
	// stat the working directory
	stat, err := os.Stat(l.WorkingDir)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func (l *LocalStorage) loadScenariosFromWorkingDir() (map[string]*Scenario, error) {
	return l.loadScenarios(l.WorkingDir)
}

func (l *LocalStorage) getCurrentGitMetadata() (*gitMetadata, error) {
	// check if the git metadata file exists
	if !l.fileExists(filepath.Join(l.WorkingDir, gitMetadataFile)) {
		return nil, fmt.Errorf("git metadata file does not exist in working directory")
	}

	data, err := os.ReadFile(filepath.Join(l.WorkingDir, gitMetadataFile))
	if err != nil {
		return nil, err
	}
	metadata := gitMetadata{}
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (l *LocalStorage) writeGitMetadata(hash, date string) error {
	metadata := gitMetadata{
		CommitHash: hash,
		Date:       date,
	}
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(l.WorkingDir, gitMetadataFile), data, 0644)
}

func (l *LocalStorage) copyScenario(scenario *Scenario) error {
	if err := copyAllScenariosFromDir(scenario.SrcDir, l.GetScenarioWorkingDir(scenario)); err != nil {
		return fmt.Errorf("unable to copy scenario to working directory:  %w", err)
	}

	stateFilePath := filepath.Join(l.GetScenarioWorkingDir(scenario), stateFile)
	if !l.fileExists(stateFilePath) {
		if err := l.WriteStateToFile(scenario, State{State: NotDeployed}); err != nil {
			return fmt.Errorf("unable to write state file to working directory:  %w", err)
		}
	}

	return nil
}

func (l *LocalStorage) createScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Pulumi.yaml: %w", err)
	}

	scenario := Scenario{}
	if err = yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse Pulumi.yaml: %w", err)
	}

	scenario.SrcDir = filepath.Dir(path)

	scenarioWorkDir := l.GetScenarioWorkingDir(&scenario)
	statePath := filepath.Join(scenarioWorkDir, stateFile)

	if !l.fileExists(statePath) {
		scenario.State.State = NotDeployed
	} else {
		data, err = os.ReadFile(statePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read state.yaml: %w", err)
		}

		if err = yaml.Unmarshal(data, &scenario.State); err != nil {
			return nil, fmt.Errorf("failed to parse state.yaml: %w", err)
		}
	}

	hashProj, err := hashFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not hash file: %w", err)
	}

	hashMain, err := hashFile(filepath.Join(filepath.Dir(path), "main.go"))
	if err != nil {
		return nil, fmt.Errorf("could not hash file: %w", err)
	}

	scenario.Hash = hashProj + hashMain
	return &scenario, nil
}

func (l *LocalStorage) fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		logrus.WithError(err).Error("unable to check if file exists")
		return false
	}

	return true
}

func (l *LocalStorage) loadScenarios(path string) (map[string]*Scenario, error) {
	scenarios := make(map[string]*Scenario)
	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error loading scenarios, unable to walk path %s: %w", path, err)
		}

		if !info.IsDir() && info.Name() == "Pulumi.yaml" {
			scenario, err := l.createScenario(path)
			if err != nil {
				return fmt.Errorf("error loading scenarios, unable to create scenario: %w", err)
			}

			scenarios[scenario.ScenarioParams.ID] = scenario
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return scenarios, nil
}

func copyAllScenariosFromDir(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	src, err := os.Open(srcDir)
	if err != nil {
		return err
	}

	defer func() {
		if err := src.Close(); err != nil {
			logrus.WithError(err).Error()
		}
	}()

	files, err := src.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(srcDir, file.Name())
		dstPath := filepath.Join(dstDir, file.Name())

		if file.IsDir() {
			if err := copyAllScenariosFromDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func getLocalWorkDirPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".cnappgoat"), nil
}

func hashFile(filepath string) (string, error) {
	// check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return "", fmt.Errorf("error when hashing the file %s, file doesn't exist: %v", filepath, err)
	}
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("error when hashing the file %s, cannot open file: %v", filepath, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			logrus.WithError(err).Error()
		}
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("error when hashing the file %s, cannot copy file: %w", filepath, err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func copyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	defer func() {
		if err := srcFile.Close(); err != nil {
			logrus.WithError(err).Error()
		}
	}()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			logrus.WithError(err).Error()
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
