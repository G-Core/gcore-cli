package config

import (
	"fmt"
	"os"
	"path"
)

const CliName = "gcore-cli"
const CliConfigFile = "config.yaml"

func getConfigHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}

	return path.Join(home, ".gcorecli"), nil
}

func GetConfigPath() (string, error) {
	configDir, err := getConfigHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(configDir, CliConfigFile), nil
}
