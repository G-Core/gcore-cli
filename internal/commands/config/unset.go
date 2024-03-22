package config

import (
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/output"
)

func unset() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "unset <key>",
		Short: "Reset property in the active profile",
		Long: "Resets property in the active profile. If property was reset the value for it will be taken from default profile.\n" +
			"The only allowed arguments are: api-url, api-key",
		ValidArgs: []string{"api-url", "api-key"},
		GroupID:   "config_commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()

				return nil
			}

			ctx := cmd.Context()
			profileName := core.ExtractProfile(ctx)
			cfg := core.ExtractConfig(ctx)
			profile, err := cfg.GetProfile(profileName)
			if err != nil {
				return err
			}

			for _, name := range args {
				switch name {
				case "api-url":
					profile.ApiUrl = nil
				case "api-key":
					profile.ApiKey = nil
				}
			}

			cfg.SetProfile(profileName, profile)
			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			if err := cfg.Save(path); err != nil {
				return err
			}

			output.Print(profile)
			return nil
		},
	}

	return cmd
}
