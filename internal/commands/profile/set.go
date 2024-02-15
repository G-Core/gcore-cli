package profile

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/errors"
)

func set() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "set <profile> <flags>",
		Short: "Set value in specific profile",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return &errors.CliError{
					Err: fmt.Errorf("profile isn't specified"),
				}
			}
			name := args[0]

			profile, found := cfg.Profiles[name]
			if !found {
				return &errors.CliError{
					Err: fmt.Errorf("profile '%s' doesn't exist", name),
				}
			}

			local, _ := cmd.PersistentFlags().GetString("local")
			url, _ := cmd.PersistentFlags().GetString("url")
			apikey, _ := cmd.PersistentFlags().GetString("apikey")
			project, _ := cmd.PersistentFlags().GetInt("project")
			region, _ := cmd.PersistentFlags().GetInt("region")

			if local == "true" {
				profile.Local = true
			}

			if url != "" {
				profile.ApiURL = url
			}

			if apikey != "" {
				profile.ApiKey = apikey
			}

			if project != 0 {
				profile.Project = project
			}

			if region != 0 {
				profile.Region = region
			}

			return nil
		},
	}

	cmd.PersistentFlags().String("local", "", "Switch CLI in dev mode")
	cmd.PersistentFlags().String("url", "", "Set url for profile")
	cmd.PersistentFlags().String("apikey", "", "Set API key for profile")
	cmd.PersistentFlags().Int("project", 0, "Set cloud project id for profile")
	cmd.PersistentFlags().Int("region", 0, "Set cloud region id for profile")

	return cmd
}
