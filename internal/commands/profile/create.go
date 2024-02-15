package profile

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
)

func create() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "create <profile_name> <flags>",
		Aliases: []string{"c"},
		Short:   "Create a profile of the config.",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := args[0]

			apikey, err := cmd.PersistentFlags().GetString("apikey")
			if err != nil {
				return err
			}

			_, ok := cfg.Profiles[profile]
			if !ok {
				cfg.Profiles[profile] = &config.Profile{}
			}
			if len(apikey) > 0 {
				cfg.Profiles[profile].ApiKey = apikey
			} else {
				fmt.Printf("Please, enter API Key: ")
				fmt.Scanf("%s\n", &apikey)
				cfg.Profiles[profile].ApiKey = apikey
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP("apikey", "", "", "Pass an API Key for profile initialization")

	return cmd
}
