package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/G-core/gcore-cli/internal/config"
)

const CliName = "gcorecli"
const CliConfigName = "config"
const CliConfigType = "yaml"
const CliEnvPrefix = "gcore"

var CliConfigPath = fmt.Sprintf("$HOME/.%s", CliName)

func SetUpDefaultConfig(v *viper.Viper) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting user home directory: %w", err)
	}

	configPath := strings.Replace(CliConfigPath, "$HOME", home, -1)
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		if err := os.Mkdir(configPath, 0766); err != nil {
			return fmt.Errorf("making dir for config: %w", err)
		}
	}

	defaultConfig := config.NewDefault()

	return defaultConfig.Save(v)
}
