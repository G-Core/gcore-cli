package subnet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
)

func displaySubnet(ctx context.Context, id string) error {
	resp, err := client.GetSubnetInstanceWithResponse(ctx, projectID, regionID, id)
	if err != nil {
		// TODO: Should we show these errors to user?
		return fmt.Errorf("failed to get subnet instance: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.ParseCloudErr(resp.Body)
	}

	output.Print(resp.JSON200)

	return nil
}

func show() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "show <id>",
		Short:             "Show information about specific subnet",
		ValidArgsFunction: core.SubnetCompletion,
		Long:              ``, // TODO: Description with examples
		Args:              cobra.ExactArgs(1),
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var networkID = args[0]

			return displaySubnet(cmd.Context(), networkID)
		},
	}

	return cmd
}
