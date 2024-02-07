package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/G-core/cli/pkg/errors"
	"github.com/G-core/cli/pkg/output"
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

			if resp.StatusCode() == http.StatusOK {
				output.Print(resp.JSON200)

				return nil
			}

			if output.IsJSON() {
				fmt.Println(string(resp.Body))

				return nil
			}

			s := struct {
				Message string `json:"message"`
			}{}

			if err := json.Unmarshal(resp.Body, &s); err != nil {
				log.Println(err)
				output.Print(err)

				return nil
			}

			output.Print(&errors.CliError{
				Err:  fmt.Errorf("%s", s.Message),
				Code: 1,
			})

			return nil
		},
	}

	return showCmd
}
