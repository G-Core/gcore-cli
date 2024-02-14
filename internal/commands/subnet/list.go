package subnet

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := client.GetSubnetWithResponse(cmd.Context(), projectID, regionID, nil)
			if err != nil {
				return fmt.Errorf("failed to get subnet list: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.ParseCloudErr(resp.Body)
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
			output.Print(subnets)

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&networkID, "network", "",
		"",
		"Display only subnets that belong to specific network")

	return cmd
}
