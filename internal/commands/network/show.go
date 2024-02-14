package network

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
)

func displayNetwork(ctx context.Context, id string) error {
	resp, err := client.GetNetworkInstanceWithResponse(ctx, projectID, regionID, id)
	if err != nil {
		// TODO: Should we show these errors to user?
		return fmt.Errorf("failed to get network instance: %w", err)
	}

	if resp.StatusCode() == http.StatusOK {
		output.Print(resp.JSON200)

		return nil
	}

	return errors.ParseCloudErr(resp.Body)
}

func show() *cobra.Command {
	// showCmd represents the create command
	var showCmd = &cobra.Command{
		Use:   "show <id>",
		Short: "Show information about specific network",
		Long:  ``, // TODO: Description with examples
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var networkID = args[0]

			return displayNetwork(cmd.Context(), networkID)
		},
	}

	return showCmd
}
