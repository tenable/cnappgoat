package infra

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
)

const (
	gitScenariosRepoURL = "https://github.com/ermetic-research/cnappgoat-scenarios.git"
)

func GitDownloadScenariosToTempDir() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "cnappgoat-scenarios-*")
	if err != nil {
		return "", "", err
	}

	r, err := gitDownload(gitScenariosRepoURL, tempDir)
	if err != nil {
		return "", "", err
	}
	logrus.Debugf("scenarios downloaded to: %s from %s", tempDir, gitScenariosRepoURL)
	headHash, err := getLatestCommitHash(r)
	if err != nil {
		return "", "", err
	}
	return tempDir, headHash, nil
}

func GitDownloadScenarios(path string) (*git.Repository, error) {
	scenariosDirectory, err := createScenariosDir(path)
	if err != nil {
		return nil, err
	}
	r, err := gitDownload(gitScenariosRepoURL, scenariosDirectory)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("scenarios downloaded to: %s from %s", scenariosDirectory, gitScenariosRepoURL)
	return r, nil
}

func createScenariosDir(path string) (string, error) {
	dir := filepath.Join(path, "CNAPPgoatScenarios")
	// Check if the directory already exists
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return "", fmt.Errorf("scenarios download directory already exists: %s", dir)
	}

	err := os.Mkdir(dir, 0755)
	if err != nil {
		return "", err
	}

	return dir, nil
}

func gitDownload(url string, path string) (*git.Repository, error) {
	r, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
		Depth:             1,
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func getLatestCommitHash(repo *git.Repository) (string, error) {
	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	// Return the commit hash
	return ref.Hash().String(), nil
}
