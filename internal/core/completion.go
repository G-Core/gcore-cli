package core

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
)

func ProfileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := cmd.Context()

	cfg := ExtractConfig(ctx)
	var completions []string
	completions = append(completions, config.DefaultProfile)
	for name, _ := range cfg.Profiles {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveDefault
}
