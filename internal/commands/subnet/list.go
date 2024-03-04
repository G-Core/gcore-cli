package subnet

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
)

func list() *cobra.Command {
	var (
		networkID string
	)

	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of client's subnets",
		Long:    ``,           // TODO: Description with examples
		Args:    cobra.NoArgs, // TODO: search by name, id etc
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
			resp, err := client.GetSubnetWithResponse(cmd.Context(), projectID, regionID, nil)
			if err != nil {
				return fmt.Errorf("failed to get subnet list: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.ParseCloudErr(resp.Body)
			}

			if resp.JSON200 == nil && len(resp.JSON200.Results) == 0 {
				output.Print("You don't have subnets")

				return nil
			}

			if networkID == "" {
				output.Print(resp.JSON200)

				return nil
			}

			subnets := make([]cloud.SubnetSchema, 0)
			for _, s := range resp.JSON200.Results {
				if s.NetworkId != networkID {
					continue
				}

				subnets = append(subnets, s)
			}

			if len(subnets) > 0 {
				output.Print(subnets)
			} else {
				output.Print("You don't have subnets")
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&networkID, "network", "",
		"",
		"Display only subnets that belong to specific network")
	cmd.RegisterFlagCompletionFunc("network", core.NetworkCompletion)

	return cmd
}
