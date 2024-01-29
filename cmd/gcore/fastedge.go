package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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
		Short: "Show list of client's apps",
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
				fmt.Printf("ID: %d\n\tStatus: %s\n\tName: %s\n\tUrl: %s\n",
					app.Id,
					statusToString(app.Status),
					app.Name,
					app.Url,
				)
			}
			return nil
		},
	}

	var cmdAppsGet = &cobra.Command{
		Use:   "get <app_id>",
		Short: "Show app details",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			rsp, err := client.GetAppWithResponse(context.Background(), id)
			if err != nil {
				return fmt.Errorf("getting the list of apps: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the list of apps: %s", string(rsp.Body))
			}
			fmt.Printf(
				"Name: %s\nBinary: %d\nPlan: %s\nStatus: %s\nUrl: %s\n",
				*(rsp.JSON200.Name),
				rsp.JSON200.Binary,
				rsp.JSON200.Plan,
				statusToString(rsp.JSON200.Status),
				*(rsp.JSON200.Url),
			)
			return nil
		},
	}

	cmdApps.AddCommand(cmdAppsList, cmdAppsGet)
	return cmdApps
}

func statusToString(s int) string {
	switch s {
	case 0:
		return "draft"
	case 1:
		return "enabled"
	case 2:
		return "disabled"
	case 3:
		return "rate limit (hourly limit)"
	case 4:
		return "rate limit (daily limit)"
	}
	return "unknown"
}
