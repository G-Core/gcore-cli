package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/G-core/cli/pkg/sdk"
	"github.com/spf13/cobra"
)

func fastedge(client *sdk.ClientWithResponses) *cobra.Command {
	var cmdFastedge = &cobra.Command{
		Use:   "fastedge <subcommand>",
		Short: "Gcore Edge compute solution",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

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
		Run: func(cmd *cobra.Command, args []string) {
			rsp, err := client.ListAppsWithResponse(context.Background())
			if err != nil {
				fmt.Printf("Cannot get the list of apps: %v\n", err)
				os.Exit(1)
			}
			if rsp.StatusCode() != http.StatusOK {
				fmt.Printf("Error getting list of apps: %v\n", string(rsp.Body))
				os.Exit(1)
			}
			if len(*rsp.JSON200) == 0 {
				fmt.Printf("you have no apps\n")
				return
			}
			for _, app := range *rsp.JSON200 {
				fmt.Printf("ID: %d\n\tStatus: %d\n\tName: %s\n\tUrl: %s\n",
					app.Id,
					app.Status,
					app.Name,
					app.Url,
				)
			}
		},
	}

	cmdFastedge.AddCommand(cmdApps)
	cmdApps.AddCommand(cmdAppsList)

	return cmdFastedge
}
