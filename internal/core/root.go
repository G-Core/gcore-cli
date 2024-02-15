package core

import (
	"context"
	"fmt"
	"github.com/G-core/gcore-cli/internal/commands/profile"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/G-core/gcore-cli/internal/commands/fastedge"
	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/human"
	"github.com/G-core/gcore-cli/internal/output"
)

func Execute() {
	var rootCmd = &cobra.Command{
		// TODO: pick name from binary name
		Use:           CliName,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// global flags, applicable to all sub-commands
	apiKey := rootCmd.PersistentFlags().StringP("apikey", "a", "", "API key")
	apiUrl := rootCmd.PersistentFlags().StringP("url", "u", "https://api.gcore.com", "API URL")
	rootCmd.PersistentFlags().BoolP("force", "f", false, `Assume answer "yes" to all "are you sure?" questions`)
	rootCmd.PersistentFlags().IntP("project", "", 0, "Cloud project ID")
	rootCmd.PersistentFlags().IntP("region", "", 0, "Cloud region ID")
	rootCmd.PersistentFlags().StringP("profile", "", "", "The config profile to use")
	rootCmd.PersistentFlags().BoolP("wait", "", false, "Wait for command result")
	rootCmd.PersistentFlags().BoolP("local", "", false, "Switch CLI to local development")
	output.FormatOption(rootCmd)
	rootCmd.ParseFlags(os.Args[1:])

	v := viper.New()

	// Config set up
	v.AddConfigPath(CliConfigPath)
	v.SetConfigName(CliConfigName)
	v.SetConfigType(CliConfigType)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := SetUpDefaultConfig(v); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create default config: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Failed to read config: %v\n", err)
			os.Exit(1)
		}
	}

	// Loading config
	var cfg config.Config
	if err := cfg.Load(v); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal config: %v\n", err)
		os.Exit(1)
	}

	v.SetEnvPrefix(CliEnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	currentProfile := getProfile(rootCmd, v, &cfg)
	rootCmd.Flags().Set("profile", currentProfile)
	bindFlags(rootCmd, v, currentProfile)

	authFunc := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "APIKey "+*apiKey)
		return nil
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		for _, safeCmd := range []string{"completion", "help"} {
			if strings.Contains(cmd.CommandPath(), safeCmd) {
				return nil
			}
		}
		if *apiUrl == "" {
			return &errors.CliError{
				Message: "URL for API isn't specified",
				Hint:    "You can specify it by -u flag or GCORE_URL env variable",
				Code:    1,
			}
		}

		return nil
	}

	fastedgeCmd, err := fastedge.Commands(*apiUrl, authFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	profileCmd, err := profile.Commands(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(fastedgeCmd)
	rootCmd.AddCommand(profileCmd)
	err = rootCmd.Execute()
	if err != nil {
		cliErr, ok := err.(*errors.CliError)
		if !ok {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}

		body, _ := human.Marshal(cliErr, nil)
		fmt.Println(body)
		os.Exit(cliErr.Code)
	}
}

func getProfile(cmd *cobra.Command, v *viper.Viper, cfg *config.Config) string {
	if cmd.Flag("profile").Changed {
		profile, _ := cmd.Flags().GetString("profile")

		return profile
	}

	if v.IsSet("profile") {
		return v.GetString("profile")
	}

	if len(cfg.CurrentProfile) != 0 {
		return cfg.CurrentProfile
	}

	return "default"
}

func bindFlags(cmd *cobra.Command, v *viper.Viper, profile string) {
	profile = fmt.Sprintf("profiles.%s", profile)
	// Apply the viper config value to the flag when the flag is not set and viper has a value
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			return
		}

		// Set as environment
		if v.IsSet(f.Name) {
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", v.Get(f.Name)))

			return
		}

		// Set in profile config
		pathInProfile := fmt.Sprintf("%s.%s", profile, f.Name)
		if v.IsSet(pathInProfile) {
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", v.Get(pathInProfile)))

			return
		}
	})
}
