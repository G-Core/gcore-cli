package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/human"
	"github.com/G-core/gcore-cli/internal/output"
)

func init() {
	cobra.EnableCommandSorting = false
}

func Execute(commands []*cobra.Command) {
	var rootCmd = &cobra.Command{
		// TODO: pick name from binary name
		Use:           os.Args[0],
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var meta meta

	// global flags, applicable to all sub-commands
	rootCmd.PersistentFlags().StringVarP(&meta.flagConfig, "config", "c", "", "The path to the config file")
	rootCmd.PersistentFlags().BoolVarP(&meta.flagForce, "force", "f", false, `Assume answer "yes" to all "are you sure?" questions`)
	rootCmd.PersistentFlags().StringVarP(&meta.flagProfile, "profile", "p", "", "The config profile to use")
	rootCmd.RegisterFlagCompletionFunc("profile", ProfileCompletion)
	rootCmd.PersistentFlags().BoolVarP(&meta.flagWait, "wait", "w", false, "Wait for command result")
	rootCmd.PersistentFlags().IntVarP(&meta.flagCloudProject, "cloud-project", "", 0, "Cloud project ID")
	rootCmd.PersistentFlags().IntVarP(&meta.flagCloudRegion, "cloud-region", "", 0, "Cloud region ID")
	rootCmd.RegisterFlagCompletionFunc("cloud-region", RegionCompletion)

	output.FormatOption(rootCmd)
	rootCmd.ParseFlags(os.Args[1:])

	meta.cfg = GetConfig()
	meta.authFunc = func(ctx context.Context, req *http.Request) error {
		profile, err := GetClientProfile(ctx)
		if err != nil {
			return err
		}

		if profile.ApiKey == nil || *profile.ApiKey == "" {
			return &errors.CliError{
				Message: "subcommand requires authorization",
				Hint:    "See gcore-cli init, gcore-cli config",
			}
		}

		req.Header.Set("Authorization", "APIKey "+*profile.ApiKey)
		return nil
	}

	meta.ctx = injectMeta(context.Background(), meta)
	rootCmd.SetContext(meta.ctx)
	rootCmd.AddGroup(&cobra.Group{
		ID:    "fastedge",
		Title: "FastEdge commands",
	}, &cobra.Group{
		ID:    "cloud",
		Title: "Cloud commands",
	}, &cobra.Group{
		ID:    "configuration",
		Title: "Configuration commands",
	})

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		for _, safeCmd := range []string{"init", "config", "completion", "help"} {
			if strings.Contains(cmd.CommandPath(), safeCmd) {
				return nil
			}
		}

		profile, err := GetClientProfile(cmd.Context())
		if err != nil {
			return err
		}

		if profile.ApiUrl == nil && *profile.ApiUrl == "" {
			return &errors.CliError{
				Message: "URL for API isn't specified",
				Hint:    "You can specify it by -u flag or GCORE_URL env variable",
				Code:    1,
			}
		}

		return nil
	}

	for _, command := range commands {
		rootCmd.AddCommand(command)
	}

	err := rootCmd.Execute()
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

// GetConfig tries to load config from $HOME dir.
// If config doesn't exist - returns default config.
func GetConfig() *config.Config {
	var (
		err error
		cfg config.Config
	)

	path := os.Getenv("GCORE_CONFIG")
	if len(path) == 0 {
		path, err = config.GetConfigPath()
		if err != nil {
			return config.NewDefault()
		}
	}

	if err := cfg.Load(path); err != nil {
		return config.NewDefault()
	}

	return &cfg
}
