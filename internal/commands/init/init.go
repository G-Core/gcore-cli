package init

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/sure"
)

func Commands() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init <flags>",
		Short: "Initialize the config for gcore-cli",
		Long: `Initialize the active profile of the config.
Default path for configuration file is based on the following priority order:
- $GCORE_CONFIG
- $HOME/.gcorecli/config.yaml
`,
		GroupID: "configuration",
		Example: "gcore init -p prod",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			profileName := core.ExtractProfile(ctx)
			cfg := core.ExtractConfig(ctx)

			var profile = &cfg.Profile
			if profileName != config.DefaultProfile {
				_, found := cfg.Profiles[profileName]
				if !found {
					if cfg.Profiles == nil {
						cfg.Profiles = make(map[string]*config.Profile)
					}
					cfg.Profiles[profileName] = &config.Profile{}
				}
				profile = cfg.Profiles[profileName]
			}

			// TODO: Interactive output should be in stderror
			if !sure.AreYou(cmd, fmt.Sprintf("overwrite profile '%s'", profileName)) {
				return errors.ErrAborted
			}

			profile.ApiKey = askForApiKey(cmd)
			client, err := makeClient(profileName, cfg)
			if err != nil {
				return err
			}

			profile.CloudProject, err = askToChooseProject(ctx, client)
			if err != nil {
				return err
			}

			profile.CloudRegion, err = askToChooseRegion(ctx, client)
			if err != nil {
				return err
			}

			path, err := core.ExtractConfigPath(ctx)
			if err != nil {
				return err
			}

			return cfg.Save(path)
		},
	}

	cmd.PersistentFlags().String("apikey", "", "GCore API key")

	return cmd
}

func makeClient(profileName string, cfg *config.Config) (*cloud.ClientWithResponses, error) {
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	var baseUrl = *profile.ApiUrl
	if !profile.IsLocal() {
		baseUrl += "/cloud"
	}

	client, err := cloud.NewClientWithResponses(baseUrl, cloud.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "APIKey "+*profile.ApiKey)

		return nil
	}))

	if err != nil {
		return nil, err
	}

	return client, nil
}

func askForApiKey(cmd *cobra.Command) *string {
	apikey, _ := cmd.PersistentFlags().GetString("apikey")
	if apikey == "" {
		fmt.Printf("Please, enter API key: ")
		fmt.Scanf("%s", &apikey)
	}

	return &apikey
}

func askToChooseProject(ctx context.Context, client *cloud.ClientWithResponses) (*int, error) {
	resp, err := client.GetProjectListWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.ParseCloudErr(resp.Body)
	}

	if resp.JSON200 == nil {
		return nil, &errors.CliError{
			Err:  fmt.Errorf("404 not found"),
			Hint: fmt.Sprintf("Check profile '%s' configuration: api-url and local", core.ExtractProfile(ctx)),
		}
	}

	if len(resp.JSON200.Results) == 1 {
		fmt.Printf("Cloud project: %d (%s)\n", resp.JSON200.Results[0].Id, resp.JSON200.Results[0].Name)

		return &resp.JSON200.Results[0].Id, nil
	}

	var defProject = resp.JSON200.Results[0]
	fmt.Printf("Please, choose default project for Cloud [%d (%s)] \n", 0, defProject.Name)
	for idx, project := range resp.JSON200.Results {
		fmt.Printf("%d - %s\n", idx, project.Name)
	}
	var idx = 0
	for {
		fmt.Scanf("%d", &idx)
		if idx < len(resp.JSON200.Results)-1 &&
			idx >= 0 {
			break
		}
	}
	fmt.Printf("Cloud project: %d (%s)\n", resp.JSON200.Results[idx].Id, resp.JSON200.Results[idx].Name)

	return &resp.JSON200.Results[idx].Id, nil
}

func askToChooseRegion(ctx context.Context, client *cloud.ClientWithResponses) (*int, error) {
	resp, err := client.GetRegionWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.ParseCloudErr(resp.Body)
	}

	if resp.JSON200 == nil {
		return nil, &errors.CliError{
			Err:  fmt.Errorf("404 not found"),
			Hint: fmt.Sprintf("Check profile '%s' configuration: api-url and local", core.ExtractProfile(ctx)),
		}
	}

	var regions []cloud.RegionSchema
	for _, region := range resp.JSON200.Results {
		if region.State == cloud.RegionSchemaStateACTIVE {
			regions = append(regions, region)
		}
	}

	if len(regions) == 0 {
		fmt.Printf("There're no available active regions")

		return nil, nil
	}

	if len(regions) == 1 {
		fmt.Printf("Cloud region: %d (%s)\n", regions[0].Id, regions[0].DisplayName)

		return &regions[0].Id, nil
	}

	var defProject = regions[0]
	fmt.Printf("Please, choose default region for Cloud [%d (%s)] \n", 0, defProject.DisplayName)
	for idx, project := range regions {
		fmt.Printf("%d - %s\n", idx, project.DisplayName)
	}

	var idx = 0
	for {
		fmt.Scanf("%d", &idx)
		if idx < len(regions)-1 &&
			idx >= 0 {
			break
		}
	}
	fmt.Printf("Cloud region: %d (%s)\n", regions[idx].Id, regions[idx].DisplayName)

	return &regions[idx].Id, nil
}
