package config

import (
	"github.com/AlekSi/pointer"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/output"
)

func dump() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "dump",
		Short:   "Dumps the config file",
		Args:    cobra.NoArgs,
		GroupID: "config_commands",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			cfg := core.ExtractConfig(ctx)

			// Secure keys
			cfg.Profile.ApiKey = pointer.To(secureKey(cfg.Profile.ApiKey))
			for _, profile := range cfg.Profiles {
				profile.ApiKey = pointer.To(secureKey(profile.ApiKey))
			}

			output.Print(cfg)

			return nil
		},
	}

	return cmd
}
