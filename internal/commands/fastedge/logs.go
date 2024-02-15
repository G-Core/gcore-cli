package fastedge

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	sdk "github.com/G-Core/FastEdge-client-sdk-go"
	"github.com/spf13/cobra"

	"github.com/G-core/cli/internal/output"
)

func appLogsFilterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "", "", "From time")
	cmd.Flags().StringP("to", "", "", "To time")
	cmd.Flags().StringP("sort", "", "asc", "Sort order")
	cmd.Flags().StringP("edge", "", "", "Edge name")
	cmd.Flags().StringP("client-ip", "", "", "Client IP")
}

// logs-related commands
func logs() *cobra.Command {
	var (
		from, to *time.Time
		sort     *sdk.GetV1AppsIdLogsParamsSort
		edge     *string
		clientIp *string
	)

	var cmdLogs = &cobra.Command{
		Use:   "logs <subcommand>",
		Short: "Logs-related commands",
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdLogsShow = &cobra.Command{
		Use:   "show <app_id>",
		Short: "Show app logs",
		Long: `Show app logs printed to stdout/stderr. 
This command allows you filtering by edge name, client ip and time range.`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			fromFlag, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}
			toFlag, err := cmd.Flags().GetString("to")
			if err != nil {
				return err
			}
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

			if fromFlag != "" {
				f, err := time.Parse(time.RFC3339, fromFlag)
				if err != nil {
					return errors.New("invalid format for `from` expected RFC3339")
				}
				from = &f
			}

			if toFlag != "" {
				t, err := time.Parse(time.RFC3339, toFlag)
				if err != nil {
					return errors.New("invalid format for `to` expected RFC3339")
				}
				to = &t
			}

			if sortFlag != "" {
				logParamSort := sdk.GetV1AppsIdLogsParamsSort(sortFlag)
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
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			rsp, err := client.GetV1AppsIdLogsWithResponse(
				context.Background(),
				id,
				&sdk.GetV1AppsIdLogsParams{
					From:     from,
					To:       to,
					Edge:     edge,
					Sort:     sort,
					ClientIp: clientIp,
				},
			)
			if err != nil {
				return fmt.Errorf("getting app logs: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting app logs: %s", string(rsp.Body))
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
				for *rsp.JSON200.CurrentPage < *rsp.JSON200.TotalPages {
					fmt.Printf("Displaying %d/%d logs, load next page? (Y/n) ", *rsp.JSON200.CurrentPage**rsp.JSON200.PageSize, *rsp.JSON200.TotalPages**rsp.JSON200.PageSize)
					text, _ := reader.ReadString('\n')
					text = strings.ToLower(strings.TrimSpace(text))

					if text != "y" {
						break
					}

					// Erase the last line
					fmt.Print("\033[2K\033[1A\033[2K\033[1A\n")

					// Increment the page number
					page := int64(*rsp.JSON200.CurrentPage + 1)

					// Call the API again with the new page number
					rsp, err = client.GetV1AppsIdLogsWithResponse(
						context.Background(),
						id,
						&sdk.GetV1AppsIdLogsParams{
							From:        from,
							To:          to,
							Edge:        edge,
							Sort:        sort,
							ClientIp:    clientIp,
							CurrentPage: &page,
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
		Use:   "enable <subcommand>",
		Short: "Enable logging",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			fmt.Printf("Enabling %d\n", id)
			return nil
		},
	}

	appLogsFilterFlags(cmdLogsShow)
	cmdLogs.AddCommand(
		cmdLogsShow,
		cmdLogEnable)

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
