package config

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/output"
)

func Commands() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "config",
		Short:   "Configuration file management",
		GroupID: "configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddGroup(&cobra.Group{
		ID:    "config_commands",
		Title: "Configuration commands",
	})

	cmd.AddCommand(info(), get(), set(), unset(), dump(), profileCmd())
	return cmd
}

func profileCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "profile",
		Short: "Commands to manage config profiles",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(listProfiles(), createProfileCmd(), switchProfileCmd(), deleteProfileCmd())
	return cmd
}

func createProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <profile>",
		Aliases: []string{"c"},
		Short:   "Create a profile",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			return []string{"api-key=", "api-url="}, cobra.ShellCompDirectiveDefault
		},

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) == 0 {
				cmd.Help()

				return nil
			}

			profileName := args[0]
			ctx := cmd.Context()
			cfg := core.ExtractConfig(ctx)

			_, exist := cfg.Profiles[profileName]
			if exist {
				return fmt.Errorf("profile '%s' already exists", profileName)
			}

			var profile = &config.Profile{}
			if len(args[1:]) != 0 {
				argProfile, err := profileFromArgs(args[1:])
				if err != nil {
					return err
				}

				profile = config.MergeProfiles(profile, argProfile)
			}

			if cfg.Profiles == nil {
				cfg.Profiles = make(map[string]*config.Profile)
			}
			cfg.Profiles[profileName] = profile
			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			return cfg.Save(path)
		},
	}

	return cmd
}

func deleteProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <profile>",
		Aliases:           []string{"d"},
		ValidArgsFunction: core.ProfileCompletion,
		Short:             "Delete profile from the config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()

				return nil
			}

			profileName := args[0]
			ctx := cmd.Context()
			cfg := core.ExtractConfig(ctx)
			active := core.ExtractProfile(ctx)

			_, exist := cfg.Profiles[profileName]
			if exist {
				delete(cfg.Profiles, profileName)
			} else {
				return fmt.Errorf("profile '%s' doesn't exist", profileName)
			}

			if active == profileName {
				cfg.ActiveProfile = config.DefaultProfile
			}

			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			return cfg.Save(path)
		},
	}

	return cmd
}

func listProfiles() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Display list of available profiles in the config",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := core.ExtractConfig(ctx)
			currentProfile := core.ExtractProfile(ctx)

			var profiles []profileView
			if currentProfile == config.DefaultProfile {
				profiles = append([]profileView{}, toProfileView("=> "+config.DefaultProfile, &cfg.Profile))
			} else {
				profiles = append([]profileView{}, toProfileView(config.DefaultProfile, &cfg.Profile))
			}

			var names []string
			for name, _ := range cfg.Profiles {
				names = append(names, name)
			}
			slices.Sort(names)

			for _, name := range names {
				var pv profileView
				if name == currentProfile {
					pv = toProfileView("=> "+name, cfg.Profiles[name])
				} else {
					pv = toProfileView(name, cfg.Profiles[name])
				}

				profiles = append(profiles, pv)
			}

			output.Print(profiles)

			return nil
		},
	}

	return cmd
}

func switchProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "switch <profile>",
		ValidArgsFunction: core.ProfileCompletion,
		Aliases:           []string{"sw"},
		Short:             "Make selected profile active",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()

				return nil
			}

			profileName := args[0]
			ctx := cmd.Context()
			cfg := core.ExtractConfig(ctx)

			_, exist := cfg.Profiles[profileName]
			if exist {
				cfg.ActiveProfile = profileName
			} else if profileName != config.DefaultProfile {
				return fmt.Errorf("profile '%s' doesn't exist", profileName)
			} else {
				cfg.ActiveProfile = config.DefaultProfile
			}

			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			return cfg.Save(path)
		},
	}

	return cmd
}
