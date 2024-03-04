package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/output"
)

func set() *cobra.Command {
	var p config.Profile
	var cmd = &cobra.Command{
		Use:   "set <argN>=<valN>",
		Short: "Set property for active profile",
		Long: "This commands overwrites the configuration file parameters with user input.\n" +
			"The only allowed arguments are: api-url, api-key, cloud-project, cloud-region",
		ValidArgs: []string{"api-url", "api-key", "cloud-project", "cloud-region"},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()

				return nil
			}

			var m = make(map[string]any)
			for _, arg := range args {
				ss := strings.Split(arg, "=")
				if len(ss) != 2 {
					continue
				}

				name, value := ss[0], ss[1]
				// TODO: reflection here
				switch name {
				case "api-url", "api-key":
					m[name] = &value
				case "cloud-project", "cloud-region":
					i, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("wrong value for '%s': %w", name, err)
					}
					m[name] = &i
				}
			}

			if len(m) == 0 {
				return fmt.Errorf("invalid arguments")
			}

			for name, value := range m {
				switch name {
				case "api-url":
					p.ApiUrl = value.(*string)
				case "api-key":
					p.ApiKey = value.(*string)
				case "cloud-project":
					p.CloudProject = value.(*int)
				case "cloud-region":
					p.CloudRegion = value.(*int)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			ctx := cmd.Context()
			cfg := core.ExtractConfig(ctx)
			profileName := core.ExtractProfile(ctx)
			profile := &cfg.Profile
			if profileName != config.DefaultProfile {
				var exist bool
				profile, exist = cfg.Profiles[profileName]
				if !exist {
					if cfg.Profiles == nil {
						cfg.Profiles = map[string]*config.Profile{}
					}
					cfg.Profiles[profileName] = &config.Profile{}
					profile = cfg.Profiles[profileName]
				}
			}

			profile = config.MergeProfiles(profile, &p)
			cfg.SetProfile(profileName, profile)

			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			if err := cfg.Save(path); err != nil {
				return err
			}

			profile, _ = cfg.GetProfile(profileName)
			profile.ApiKey = pointer.To(secureKey(profile.ApiKey))
			output.Print(profile)

			return nil
		},
	}

	return cmd
}
