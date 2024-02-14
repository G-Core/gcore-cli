package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/G-core/cli/internal/commands/fastedge"
	"github.com/G-core/cli/internal/errors"
	"github.com/G-core/cli/internal/human"
	"github.com/G-core/cli/internal/output"
)

func Execute() {
	var rootCmd = &cobra.Command{
		// TODO: pick name from binary name
		Use:           "gcore-cli",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// global flags, applicable to all sub-commands
	apiKey := rootCmd.PersistentFlags().StringP("apikey", "a", "", "API key")
	apiUrl := rootCmd.PersistentFlags().StringP("url", "u", "https://api.gcore.com", "API URL")
	rootCmd.PersistentFlags().BoolP("force", "f", false, `Assume answer "yes" to all "are you sure?" questions`)
	rootCmd.PersistentFlags().IntP("project", "", 0, "Cloud project ID")
	rootCmd.PersistentFlags().IntP("region", "", 0, "Cloud region ID")
	rootCmd.PersistentFlags().BoolP("wait", "", false, "Wait for command result")
	output.FormatOption(rootCmd)
	rootCmd.ParseFlags(os.Args[1:])

	v := viper.New()
	v.SetEnvPrefix("gcore")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	bindFlags(rootCmd, v)

	authFunc := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "APIKey "+*apiKey)
		return nil
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if *apiUrl == "" {
			return &errors.CliError{
				Message: "URL for API isn't specified",
				Hint:    "You can specify it by -u flag or GCORE_URL env variable",
				Code:    1,
			}
		}

		if *apiKey == "" {
			return &errors.CliError{
				Message: "API key must be specified",
				Hint: "You can specify it with -a flag or GCORE_APIKEY env variable.\n" +
					"To get an APIKEY visit https://accounts.gcore.com/profile/api-tokens",
				Code: 1,
			}
		}

		return nil
	}

	fastedgeCmd, err := fastedge.Commands(*apiUrl, authFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(fastedgeCmd)
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

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
