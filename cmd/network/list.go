package network

import (
	"fmt"
	"github.com/G-core/cli/pkg/human"
	"github.com/G-core/cli/pkg/output"
	"github.com/spf13/cobra"
	"net/http"
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

			if resp.StatusCode() != http.StatusOK {
				// TODO: process errors from the server
				return fmt.Errorf("getting the list of networks: %s", string(resp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(resp.Body))
				return nil
			}

			var body string
			if resp.JSON200 == nil {
				body, err = human.Marshal(nil, &human.MarshalOpt{
					Title: "Networks",
				})
			} else {
				// TODO: Automate this. At least we need to show ID first.
				body, err = human.Marshal(resp.JSON200.Results, &human.MarshalOpt{
					Title: "Networks",
					Fields: []*human.MarshalFieldOpt{
						{
							FieldName: "Id",
							Label:     "ID",
						},
						{
							FieldName: "Name",
						},
						{
							FieldName: "Type",
						},
						{
							FieldName: "Shared",
						},
						{
							FieldName: "Mtu",
						},
						{
							FieldName: "Default",
						},
						{
							FieldName: "Subnets",
						},
						{
							FieldName: "Metadata",
						},
						{
							FieldName: "External",
						},
						{
							FieldName: "CreatedAt",
						},
						{
							FieldName: "UpdatedAt",
						},
						{
							FieldName: "ProjectId",
							Label:     "Project",
						},
						{
							FieldName: "Region",
						},
						{
							FieldName: "RegionId",
						},
					},
				})
			}

			if err != nil {
				return fmt.Errorf("failed to marshal to human: %w", err)
			}

			fmt.Println(body)

			return nil
		},
	}

	return listCmd
}
