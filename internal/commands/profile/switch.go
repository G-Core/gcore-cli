package profile

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/errors"
)

func switchTo() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "switch <profile>",
		Short: "Switch configuration profile",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			_, found := cfg.Profiles[name]
			if !found {
				return &errors.CliError{
					Err: fmt.Errorf("profile '%s' doesn't exist", name),
				}
			}

			cfg.CurrentProfile = name

			return nil
		},
	}

	return cmd
}
