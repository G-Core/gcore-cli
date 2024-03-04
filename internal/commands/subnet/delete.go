package subnet

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
	"github.com/G-core/gcore-cli/internal/terminal"
)

func deleteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "delete <id>",
		Aliases:           []string{"d"},
		Short:             "Delete a specific subnet",
		Long:              ``,
		ValidArgsFunction: core.SubnetCompletion,
		Args:              cobra.MinimumNArgs(1),
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

			resp, err := client.DeleteSubnetInstanceWithResponse(cmd.Context(), projectID, regionID, networkId)
			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.ParseCloudErr(resp.Body)
			}

			if !waitForResult {
				// TODO: Use a logger which writes to stderr instead of fmt
				output.Print(fmt.Sprintf("Deleting network: %s", resp.JSON200.Tasks[0]))
				return nil
			}

			taskID := resp.JSON200.Tasks[0]

			_, err = cloud.WaitForStatus(cmd.Context(), client, taskID, cloud.TaskSchemaDetailedStateFINISHED, time.Second*5, true)
			if err != nil {
				return err
			}

			// TODO: Good message are green, bad message are red etc
			fmt.Println(terminal.Style(fmt.Sprintf("Subnet '%s' deleted", networkId), color.FgGreen))

			return nil
		},
	}

	return cmd
}
