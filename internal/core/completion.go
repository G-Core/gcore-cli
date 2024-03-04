package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
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
	var completions []string

	ctx := cmd.Context()

	client, err := CloudClient(ctx)
	if err != nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	resp, err := client.GetRegionWithResponse(ctx, nil)
	if err != nil || resp.JSON200 == nil {
		return completions, cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveDefault
	}

	var rm = make(map[int]string)
	var ids []int
	for _, item := range resp.JSON200.Results {
		if item.State == cloud.RegionSchemaStateACTIVE {
			rm[item.Id] = item.DisplayName
			ids = append(ids, item.Id)
		}
	}
	slices.Sort(ids)

	completions = make([]string, len(rm))
	for i, id := range ids {
		completions[i] = fmt.Sprintf("%d\t%s", id, rm[id])
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
