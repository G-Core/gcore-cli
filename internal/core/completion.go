package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
)

func profileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := cmd.Context()

	cfg := ExtractConfig(ctx)
	var competitions []string
	competitions = append(competitions, config.DefaultProfile)
	for name, _ := range cfg.Profiles {
		if strings.HasPrefix(name, toComplete) {
			competitions = append(competitions, name)
		}
	}

	return competitions, cobra.ShellCompDirectiveDefault
}

func regionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var competitions = make([]string, len(Regions))
	var ids []int
	for id, _ := range Regions {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	for idx, id := range ids {
		competitions[idx] = fmt.Sprintf("%d\t%s", id, Regions[id])
	}

	return competitions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
}
