package profile

import (
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/output"
)

type view struct {
	Name string
	config.Profile
}

func list() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List of configs profiles",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var profiles []view

			for name, profile := range cfg.Profiles {
				profiles = append(profiles, view{Name: name, Profile: *profile})
			}
			output.Print(profiles)

			return nil
		},
	}

	return cmd
}
