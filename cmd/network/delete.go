package network

import (
	"fmt"
	"net/http"
	"time"

	"github.com/G-core/cli/pkg/terminal"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/cli/pkg/errors"
)

func delete() *cobra.Command {
	// deleteCmd represents the create command
	var deleteCmd = &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"d"},
		Short:   "Delete a specific network",
		Long:    ``,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Autogenerate names for networks as option
			networkId := args[0]

			resp, err := client.DeleteNetworkInstanceWithResponse(cmd.Context(), projectID, regionID, networkId)
			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.ParseCloudErr(resp.Body)
			}

			if !waitForResult {
				// TODO: Use a logger which writes to stderr instead of fmt
				fmt.Printf("Deleting network '%s'", networkId)
				return nil
			}

			taskID := resp.JSON200.Tasks[0]

			_, err = cloud.WaitForStatus(cmd.Context(), client, taskID, cloud.TaskSchemaDetailedStateFINISHED, time.Second*5, true)
			if err != nil {
				return err
			}

			// TODO: Good message are green, bad message are red etc
			fmt.Println(terminal.Style(fmt.Sprintf("Network '%s' deleted", networkId), color.FgGreen))

			return nil
		},
	}

	return deleteCmd
}
