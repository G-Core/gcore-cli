package network

import (
	"net/http"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
)

func rename() *cobra.Command {
	// renameCmd represents the create command
	var renameCmd = &cobra.Command{
		Use:               "rename <id> <new-name>",
		Short:             "Rename a specific network",
		ValidArgsFunction: core.NetworkCompletion,
		Long:              ``,
		Args:              cobra.MinimumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			projectID, err = core.ExtractCloudProject(ctx)
			if err != nil {
				return err
			}

			regionID, err = core.ExtractCloudRegion(ctx)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Autogenerate names for networks as option
			networkId := args[0]
			name := args[1]

			if err := validateNetworkName(name); err != nil {
				return err
			}

			resp, err := client.PatchNetworkInstanceWithResponse(cmd.Context(), projectID, regionID, networkId, cloud.NameSchema{
				Name: name,
			})

			if err != nil {
				return err
			}

			if resp.StatusCode() == http.StatusOK {
				output.Print(resp.JSON200)

				return nil
			}

			return errors.ParseCloudErr(resp.Body)
		},
	}

	return renameCmd
}
