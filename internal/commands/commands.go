package commands

import (
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/commands/config"
	"github.com/G-core/gcore-cli/internal/commands/fastedge"
	initCmd "github.com/G-core/gcore-cli/internal/commands/init"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{
		fastedge.Commands(),
		initCmd.Commands(),
		config.Commands(),
	}
}
