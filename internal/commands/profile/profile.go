package profile

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/G-core/gcore-cli/internal/config"
)

var cfg config.Config

func Commands(v *viper.Viper) (*cobra.Command, error) {
	var cmd = &cobra.Command{
		Use:   "profile <subcommand>",
		Short: "Commands to work with configuration",
		Long:  "See also gcore-cli init",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Load(v); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return cfg.Save(v)
		},
	}

	cmd.AddCommand(create(), switchTo(), set(), show(), list())

	return cmd, nil
}
