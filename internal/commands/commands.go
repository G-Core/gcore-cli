package commands

import (
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/commands/config"
	"github.com/G-core/gcore-cli/internal/commands/fastedge"
	initCmd "github.com/G-core/gcore-cli/internal/commands/init"
	"github.com/G-core/gcore-cli/internal/commands/network"
	"github.com/G-core/gcore-cli/internal/commands/subnet"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{
		fastedge.Commands(),
		network.Commands(),
		subnet.Commands(),
		initCmd.Commands(),
		config.Commands(),
	}
}
