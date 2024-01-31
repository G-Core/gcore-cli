package main

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

	"github.com/G-core/cli/pkg/sdk"
)

func main() {
	var rootCmd = &cobra.Command{Use: "gcore"}

	// global flags, applicable to all sub-commands
	apiKey := rootCmd.PersistentFlags().StringP("apikey", "a", "", "API key")
	apiUrl := rootCmd.PersistentFlags().StringP("url", "u", "https://api.gcore.com", "API URL")
	formatOption(rootCmd)
	rootCmd.ParseFlags(os.Args[1:])

	v := viper.New()
	v.SetEnvPrefix("gcore")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	bindFlags(rootCmd, v)

	client, err := sdk.NewClientWithResponses(
		*apiUrl,
		sdk.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "APIKey "+*apiKey)
			return nil
		}),
	)
	if err != nil {
		fmt.Printf("Cannot init the client: %v\n", err)
		os.Exit(1)
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

	rootCmd.AddCommand(fastedge(client))
	err = rootCmd.Execute()
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
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
