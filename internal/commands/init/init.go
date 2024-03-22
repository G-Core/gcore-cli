package init

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/sure"
)

func Commands() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init <flags>",
		Short: "Initialize the config for gcore-cli",
		Long: `Initialize the active profile of the config.
Default path for configuration file is based on the following priority order:
- $GCORE_CONFIG
- $HOME/.gcorecli/config.yaml
`,
		GroupID: "configuration",
		Example: "gcore init -p prod",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			profileName := core.ExtractProfile(ctx)
			cfg := core.ExtractConfig(ctx)

			var profile = &cfg.Profile
			if profileName != config.DefaultProfile {
				_, found := cfg.Profiles[profileName]
				if !found {
					if cfg.Profiles == nil {
						cfg.Profiles = make(map[string]*config.Profile)
					}
					cfg.Profiles[profileName] = &config.Profile{}
				}
				profile = cfg.Profiles[profileName]
			}

			// Do not ask if it's default profile, and it isn't initialized
			if profileName != config.DefaultProfile || profile.IsInitialized() {
				// TODO: Interactive output should be in stderror
				if !sure.AreYou(cmd, fmt.Sprintf("overwrite profile '%s'", profileName)) {
					return errors.ErrAborted
				}
			}

			profile.ApiKey = askForApiKey(cmd)
			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			return cfg.Save(path)
		},
	}

	cmd.PersistentFlags().String("apikey", "", "GCore API key")

	return cmd
}

func askForApiKey(cmd *cobra.Command) *string {
	apikey, _ := cmd.PersistentFlags().GetString("apikey")
	if apikey == "" {
		fmt.Printf("Please, enter API key: ")
		fmt.Scanf("%s", &apikey)
	}

	return &apikey
}
