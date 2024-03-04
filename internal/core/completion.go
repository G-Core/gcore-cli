package core

import (
	"fmt"
	"slices"
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

func RegionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions = make([]string, len(Regions))
	var ids []int
	for id, _ := range Regions {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	for idx, id := range ids {
		completions[idx] = fmt.Sprintf("%d\t%s", id, Regions[id])
	}

	return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
}

func NetworkCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	ctx := cmd.Context()

	client, err := CloudClient(ctx)
	if err != nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	profile, err := GetClientProfile(ctx)
	if err != nil || profile.CloudRegion == nil || profile.CloudProject == nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	resp, err := client.GetNetworkWithResponse(ctx, *profile.CloudProject, *profile.CloudRegion, nil)
	if err != nil || resp.JSON200 == nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	for _, item := range resp.JSON200.Results {
		if strings.HasPrefix(item.Id, toComplete) {
			completions = append(completions, fmt.Sprintf("%s\t%s", item.Id, item.Name))
		}
	}

	return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
}

func SubnetCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string

	ctx := cmd.Context()

	client, err := CloudClient(ctx)
	if err != nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	profile, err := GetClientProfile(ctx)
	if err != nil || profile.CloudRegion == nil || profile.CloudProject == nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	resp, err := client.GetSubnetWithResponse(ctx, *profile.CloudProject, *profile.CloudRegion, nil)
	if err != nil || resp.JSON200 == nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	for _, item := range resp.JSON200.Results {
		if strings.HasPrefix(item.Id, toComplete) {
			completions = append(completions, fmt.Sprintf("%s\t%s", item.Id, item.Name))
		}
	}

	return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
}
