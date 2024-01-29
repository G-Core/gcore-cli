package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/G-core/cli/pkg/sdk"
	"github.com/spf13/cobra"
)

// top-level FastEdge command
func fastedge(client *sdk.ClientWithResponses) *cobra.Command {
	var cmdFastedge = &cobra.Command{
		Use:   "fastedge <subcommand>",
		Short: "Gcore Edge compute solution",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	cmdFastedge.AddCommand(apps(client))

	return cmdFastedge
}

// apps-related commands
func apps(client *sdk.ClientWithResponses) *cobra.Command {
	var cmdApps = &cobra.Command{
		Use:   "apps <subcommand>",
		Short: "App-related commands",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdAppsList = &cobra.Command{
		Use:   "list",
		Short: "Show list of client's applications",
		Long:  ``,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.ListAppsWithResponse(context.Background())
			if err != nil {
				return fmt.Errorf("getting the list of apps: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the list of apps: %s", string(rsp.Body))
			}
			if len(*rsp.JSON200) == 0 {
				fmt.Printf("you have no apps\n")
				return nil
			}
			for _, app := range *rsp.JSON200 {
				fmt.Printf("ID: %d\n\tStatus: %d\n\tName: %s\n\tUrl: %s\n",
					app.Id,
					app.Status,
					app.Name,
					app.Url,
				)
			}
			return nil
		},
	}

	cmdApps.AddCommand(cmdAppsList)
	return cmdApps
}
