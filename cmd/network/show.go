package network

import (
	"fmt"
	"net/http"

	"github.com/G-core/cli/pkg/human"
	"github.com/G-core/cli/pkg/output"
	"github.com/spf13/cobra"
)

func show() *cobra.Command {
	// showCmd represents the create command
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show information about specific network",
		Long:  ``, // TODO: Description with examples
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			var networkID = args[0]
			resp, err := client.GetNetworkInstanceWithResponse(cmd.Context(), projectID, regionID, networkID)
			if err != nil {
				return fmt.Errorf("failed to get network instance: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				// TODO: process errors from the server
				return fmt.Errorf("getting the network instance: %s", string(resp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(resp.Body))
				return nil
			}

			body, err := human.Marshal(resp.JSON200, &human.MarshalOpt{
				Title: "Network instance",
			})
			if err != nil {
				return fmt.Errorf("failed to marshal data to 'human' format: %w", err)
			}

			fmt.Println(body)

			return nil
		},
	}

	return showCmd
}
