package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/G-core/cli/cmd/fastedge"
	"github.com/G-core/cli/cmd/network"
	"github.com/G-core/cli/pkg/output"
)

func Execute() {
	var rootCmd = &cobra.Command{
		Use:          "gcore",
		SilenceUsage: true,
	}

	// global flags, applicable to all sub-commands
	apiKey := rootCmd.PersistentFlags().StringP("apikey", "a", "", "API key")
	apiUrl := rootCmd.PersistentFlags().StringP("url", "u", "https://api.gcore.com", "API URL")
	rootCmd.PersistentFlags().BoolP("force", "f", false, `Assume answer "yes" to all "are you sure?" questions`)
	rootCmd.PersistentFlags().IntP("project", "", 0, "Cloud project ID")
	rootCmd.PersistentFlags().IntP("region", "", 0, "Cloud region ID")
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
			return errors.New("URL must be specified either with -u flag or GCORE_URL env var")
		}
		if *apiKey == "" {
			return errors.New("API Key must be specified either with -a flag or GCORE_APIKEY env var")
		}

		return nil
	}

	fastedgeCmd, err := fastedge.Commands(*apiUrl, authFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	networkCmd, err := network.Commands(*apiUrl, authFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(fastedgeCmd)
	rootCmd.AddCommand(networkCmd)
	err = rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
		os.Exit(1)
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
