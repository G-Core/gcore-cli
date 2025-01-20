package fastedge

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	sdk "github.com/G-Core/FastEdge-client-sdk-go"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/output"
)

func appLogsFilterFlags(cmd *cobra.Command) {
	cmd.Flags().String("from", "today", "Reporting period start, UTC")
	cmd.Flags().String("to", "now", "Reporting period end, UTC")
	cmd.Flags().String("sort", "asc", "Log sort order, asc or desc")
	cmd.Flags().String("edge", "", "Edge name filter")
	cmd.Flags().String("client-ip", "", "Client IP filter")
	cmd.Flags().MarkHidden("client-ip")
}

// logs-related commands
func logs() *cobra.Command {
	var (
		from, to time.Time
		sort     *sdk.ListLogsParamsSort
		edge     *string
		clientIp *string
	)

	var cmdLogs = &cobra.Command{
		Use:   "logs <subcommand>",
		Short: "Logs-related commands",
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdLogsShow = &cobra.Command{
		Use:   "show <app_name>",
		Short: "Show app logs",
		Long: `Show app logs printed to stdout/stderr. 
This command allows you filtering by edge name, client ip and time range.`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {

			sortFlag, err := cmd.Flags().GetString("sort")
			if err != nil {
				return err
			}
			edgeFlag, err := cmd.Flags().GetString("edge")
			if err != nil {
				return err
			}
			clientIpFlag, err := cmd.Flags().GetString("client-ip")
			if err != nil {
				return err
			}

			from, err = parseTimeFlag(cmd, "from")
			if err != nil {
				return fmt.Errorf("cannot parse 'from' time: %w", err)
			}

			to, err = parseTimeFlag(cmd, "to")
			if err != nil {
				return fmt.Errorf("cannot parse 'to' time: %w", err)
			}

			if sortFlag != "" {
				logParamSort := sdk.ListLogsParamsSort(sortFlag)
				if logParamSort != sdk.ListLogsParamsSortAsc && logParamSort != sdk.ListLogsParamsSortDesc {
					return errors.New("invalid value for `sort` expected asc or desc")
				}
				sort = &logParamSort
			}

			if edgeFlag != "" {
				edge = &edgeFlag
			}

			if clientIpFlag != "" {
				clientIp = &clientIpFlag
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getAppIdByName(args[0])
			if err != nil {
				return fmt.Errorf("cannot find app by name: %w", err)
			}

			rsp, err := client.ListLogsWithResponse(
				context.Background(),
				id,
				&sdk.ListLogsParams{
					From:     &from,
					To:       &to,
					Edge:     edge,
					Sort:     sort,
					ClientIp: clientIp,
				},
			)
			if err != nil {
				return fmt.Errorf("getting app logs: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting app logs: %s", extractErrorMessage(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if rsp.JSON200 == nil || rsp.JSON200.Logs == nil || len(*rsp.JSON200.Logs) == 0 {
				fmt.Printf("No logs found\n")
				return nil
			}

			reader := bufio.NewReader(os.Stdin)

			if rsp.JSON200.Logs != nil {
				printLogs(rsp.JSON200.Logs)
				for *rsp.JSON200.Offset < *rsp.JSON200.TotalCount {
					fmt.Printf("Displaying %d/%d logs, load next page? (Y/n) ", *rsp.JSON200.Offset, *rsp.JSON200.TotalCount)
					text, _ := reader.ReadString('\n')
					text = strings.ToLower(strings.TrimSpace(text))

					if text != "y" {
						break
					}

					// Erase the last line
					fmt.Print("\033[2K\033[1A\033[2K\033[1A\n")

					// Increment the page number
					var (
						offset = int32(*rsp.JSON200.Offset + 25)
						limit  = int32(25)
					)

					// Call the API again with the new page number
					rsp, err = client.ListLogsWithResponse(
						context.Background(),
						id,
						&sdk.ListLogsParams{
							From:     &from,
							To:       &to,
							Edge:     edge,
							Sort:     sort,
							ClientIp: clientIp,
							Offset:   &offset,
							Limit:    &limit,
						},
					)
					if err != nil {
						fmt.Printf("Error getting next page of logs: %v\n", err)
						break
					}

					// Print the logs from the new page
					if rsp.JSON200.Logs != nil {
						printLogs(rsp.JSON200.Logs)
					}
				}
			}
			return nil
		},
	}

	var cmdLogEnable = &cobra.Command{
		Use:   "enable <app_name>",
		Short: "Enable app logging",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getAppIdByName(args[0])
			if err != nil {
				return fmt.Errorf("cannot find app by name: %w", err)
			}
			rsp, err := client.PatchAppWithResponse(
				context.Background(),
				id,
				sdk.App{Debug: newPointer(true)},
			)
			if err != nil {
				return fmt.Errorf("enabling logging: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("enabling logging: %s", extractErrorMessage(rsp.Body))
			}

			rsp1, err := client.GetAppWithResponse(
				context.Background(),
				id,
			)
			if err != nil {
				return fmt.Errorf("getting app detail: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting app details: %s", extractErrorMessage(rsp.Body))
			}

			if rsp1.JSON200.DebugUntil == nil {
				return fmt.Errorf("logging not enabled")
			}

			fmt.Printf("Logging for app %d enabled until %v\n", id, *rsp1.JSON200.DebugUntil)
			return nil
		},
	}

	var cmdLogDisable = &cobra.Command{
		Use:   "disable <app_name>",
		Short: "Disable app logging",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getAppIdByName(args[0])
			if err != nil {
				return fmt.Errorf("cannot find app by name: %w", err)
			}
			rsp, err := client.PatchAppWithResponse(
				context.Background(),
				id,
				sdk.App{Debug: newPointer(false)},
			)
			if err != nil {
				return fmt.Errorf("disabling logging: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("disabling logging: %s", extractErrorMessage(rsp.Body))
			}

			fmt.Printf("Logging for app %d disabled\n", id)
			return nil
		},
	}

	appLogsFilterFlags(cmdLogsShow)
	cmdLogs.AddCommand(
		cmdLogsShow,
		cmdLogEnable,
		cmdLogDisable,
	)

	return cmdLogs
}

func printLogs(logs *[]sdk.Log) {
	if logs != nil {
		for _, log := range *logs {
			// Ensure pointers are not nil before dereferencing
			timestamp := ""
			if log.Timestamp != nil {
				timestamp = log.Timestamp.String()
			}

			edge := ""
			if log.Edge != nil {
				edge = *log.Edge
			}

			clientIp := ""
			if log.ClientIp != nil {
				clientIp = *log.ClientIp
			}

			logMsg := ""
			if log.Log != nil {
				logMsg = *log.Log
			}

			fmt.Printf("%s [%s] [%s] %s\n", timestamp, edge, clientIp, logMsg)
		}
	}
}
