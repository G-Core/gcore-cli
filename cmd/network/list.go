package network

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/G-core/cli/pkg/errors"
	"github.com/G-core/cli/pkg/output"
	"github.com/spf13/cobra"
)

func list() *cobra.Command {
	// listCmd represents the create command
	var listCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of client's networks",
		Long:    ``,           // TODO: Description with examples
		Args:    cobra.NoArgs, // TODO: search by name, id etc
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			resp, err := client.GetNetworkWithResponse(cmd.Context(), projectID, regionID, nil)
			if err != nil {
				return fmt.Errorf("failed to get network list: %w", err)
			}

			if resp.StatusCode() == http.StatusOK {
				output.Print(resp.JSON200.Results)

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
				output.Print(err)

				return nil
			}

			output.Print(&errors.CliError{
				Message: s.Message,
				Code:    1,
			})

			return nil
		},
	}

	return listCmd
}
