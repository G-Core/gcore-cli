package profile

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
)

func show() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <profile>",
		Short: "Displays info about profile",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			profile, found := cfg.Profiles[name]
			if found {
				output.Print(profile)

				return nil
			}

			return &errors.CliError{
				Message: fmt.Sprintf("Profile '%s' doesn't exist", name),
			}
		},
	}

	return cmd
}
