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

	e "github.com/G-core/cli/internal/errors"
	"github.com/G-core/cli/internal/output"
	"github.com/G-core/cli/internal/sure"
)

// app-related commands
func app() *cobra.Command {
	var cmdApp = &cobra.Command{
		Use:   "app <subcommand>",
		Short: "App-related commands",
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "deploy"},
		Short:   "Add new app",
		Long: `Add new FastEdge app, specifying app properties using flags.
By default, unless --disabled is specified, app is automatically deployed on all edges.
You can use either previously-uploaded binary, by specifying "--binary <id>", or
uploading binary using "--file <filename>". To load file from stdin, use "-" as filename`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := parseAppProperties(cmd)
			if err != nil {
				return err
			}
			if app.Plan == nil {
				return errors.New("plan must be specified")
			}
			if app.Binary == nil {
				file, err := cmd.Flags().GetString("file")
				if err != nil {
					return fmt.Errorf("cannot parse file name: %w", err)
				}
				if file == "" {
					return errors.New("binary must be specified either using --binary <id> or --file <filename>")
				}
				id, err := uploadBinary(file)
				if err != nil {
					return err
				}
				app.Binary = &id
			}

			rsp, err := client.AddAppWithResponse(context.Background(), app)
			if err != nil {
				return fmt.Errorf("adding the app: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("adding the app: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf(
				"ID:\t%d\nName:\t%s\nStatus:\t%s\nUrl:\t%s\n",
				rsp.JSON200.Id,
				rsp.JSON200.Name,
				appStatusToString(rsp.JSON200.Status),
				rsp.JSON200.Url,
			)
			return nil
		},
	}
	appPropertiesFlags(cmdCreate)

	var cmdUpdate = &cobra.Command{
		Use:   "update <app_id>",
		Short: "Update the app",
		Long: `This command allows to change only specified properties of the app,
omitted properties are left intact. When changing key-value properties, such
as 'env' and 'rsp_headers', new keys are added to the list, existing keys are
updated, keys with empty values are deleted.
You can use either previously-uploaded binary, by specifying "--binary <id>", or
uploading binary using "--file <filename>". To load file from stdin, use "-" as filename`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}

			app, err := parseAppProperties(cmd)
			if err != nil {
				return err
			}
			if app.Binary == nil {
				file, err := cmd.Flags().GetString("file")
				if err != nil {
					return fmt.Errorf("cannot parse file name: %w", err)
				}
				if file != "" {
					id, err := uploadBinary(file)
					if err != nil {
						return err
					}
					app.Binary = &id
				}
			}

			if !sure.AreYou(cmd, fmt.Sprintf("update app %d", id)) {
				return e.ErrAborted
			}

			rsp, err := client.PatchAppWithResponse(context.Background(), id, app)
			if err != nil {
				return fmt.Errorf("updating the app: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("updating the app: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf(
				"ID:\t%d\nName:\t%s\nStatus:\t%s\nUrl:\t%s\n",
				rsp.JSON200.Id,
				rsp.JSON200.Name,
				appStatusToString(rsp.JSON200.Status),
				rsp.JSON200.Url,
			)
			return nil
		},
	}
	appPropertiesFlags(cmdUpdate)

	var cmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of client's apps",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.ListAppsWithResponse(context.Background())
			if err != nil {
				return fmt.Errorf("getting the list of apps: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the list of apps: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if len(*rsp.JSON200) == 0 {
				fmt.Printf("you have no apps\n")
				return nil
			}

			table := make([][]string, len(*rsp.JSON200)+1)
			table[0] = []string{"ID", "Status", "Name", "Url"}
			for i, app := range *rsp.JSON200 {
				table[i+1] = []string{
					strconv.FormatInt(app.Id, 10),
					appStatusToString(app.Status),
					app.Name,
					app.Url,
				}
			}
			output.Table(table, output.Format(cmd))
			return nil
		},
	}

	var cmdGet = &cobra.Command{
		Use:     "show <app_id>",
		Aliases: []string{"get"},
		Short:   "Show app details",
		Long: `Show app properties. This command doesn't show app call statisrics.
To see statistics, use "fastedge stat app_calls" and "fastedge stat app_duration"
commands.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			rsp, err := client.GetAppWithResponse(
				context.Background(),
				id,
			)
			if err != nil {
				return fmt.Errorf("getting app detail: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting app details: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf(
				"Name:\t%s\nBinary:\t%d\nPlan:\t%s\nStatus:\t%s\nUrl:\t%s\n",
				*rsp.JSON200.Name,
				*rsp.JSON200.Binary,
				*rsp.JSON200.Plan,
				appStatusToString(*rsp.JSON200.Status),
				*rsp.JSON200.Url,
			)
			outputMap(rsp.JSON200.Env, "Env")
			outputMap(rsp.JSON200.RspHeaders, "Response headers")
			return nil
		},
	}

	var cmdEnable = &cobra.Command{
		Use:   "enable <app_id>",
		Short: "Enable the app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			rsp, err := client.PatchAppWithResponse(
				context.Background(),
				id,
				sdk.App{Status: newPointer(1)},
			)
			if err != nil {
				return fmt.Errorf("enabling app: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("enabling app: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf("App %d enabled\n", id)
			return nil
		},
	}

	var cmdDisable = &cobra.Command{
		Use:   "disable <app_id>",
		Short: "Disable the app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			rsp, err := client.PatchAppWithResponse(
				context.Background(),
				id,
				sdk.App{Status: newPointer(2)},
			)
			if err != nil {
				return fmt.Errorf("disabling app: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("disabling app: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf("App %d disabled\n", id)
			return nil
		},
	}

	var cmdDelete = &cobra.Command{
		Use:     "delete <app_id>",
		Short:   "Delete the app",
		Aliases: []string{"rm"},
		Long: `This command deletes the app. The binary, referenced by the app, is not deleted,
however binaries, not referenced by any app, get deleted by cleanup process regularly,
so if you don't want this to happen, consider disabling the app to keep binary referenced`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}

			if !sure.AreYou(cmd, fmt.Sprintf("delete app %d", id)) {
				return e.ErrAborted
			}

			rsp, err := client.DelAppWithResponse(context.Background(), id)
			if err != nil {
				return fmt.Errorf("deleting app: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("deleting app: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf("App %d deleted\n", id)
			return nil
		},
	}

	var (
		from, to *time.Time
		sort     *sdk.GetV1AppsIdLogsParamsSort
		edge     *string
		clientIp *string
	)

	var cmdLogs = &cobra.Command{
		Use:   "logs <app_id>",
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

			if rsp.JSON200.Logs == nil || len(*rsp.JSON200.Logs) == 0 {
				fmt.Printf("No logs found\n")
				return nil
			}

			reader := bufio.NewReader(os.Stdin)

			if rsp.JSON200.Logs != nil {
				printLogs(rsp.JSON200.Logs)
				for *rsp.JSON200.CurrentPage < *rsp.JSON200.TotalPages {
					fmt.Printf("Load next page? (Y/n): ")
					text, _ := reader.ReadString('\n')
					text = strings.ToLower(strings.TrimSpace(text))

					if text != "y" {
						fmt.Println("bye...")
						break
					}

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
					fmt.Println(*rsp.JSON200.CurrentPage, *rsp.JSON200.TotalPages)
				}
			}
			return nil
		},
	}
	appLogsFilterFlags(cmdLogs)

	cmdApp.AddCommand(
		cmdList,
		cmdGet,
		cmdEnable,
		cmdDisable,
		cmdCreate,
		cmdUpdate,
		cmdDelete,
		cmdLogs,
	)
	return cmdApp
}

func appPropertiesFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "App name")
	cmd.Flags().Int64("binary", 0, "Wasm binary id")
	cmd.Flags().String("file", "", "Wasm binary filename ('-' means stdin)")
	cmd.Flags().String("plan", "", "Plan name")
	cmd.Flags().Bool("disabled", false, "Set status to 'disabled'")
	cmd.Flags().StringSlice("env", nil, "Environment, in name=value format")
	cmd.Flags().StringSlice("rsp_headers", nil, "Response headers to add, in name=value format")
}

func appLogsFilterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "", "", "From time")
	cmd.Flags().StringP("to", "", "", "To time")
	cmd.Flags().StringP("sort", "", "asc", "Sort order")
	cmd.Flags().StringP("edge", "", "", "Edge name")
	cmd.Flags().StringP("client-ip", "", "", "Client IP")
}

func parseAppProperties(cmd *cobra.Command) (sdk.App, error) {
	var app sdk.App

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return app, err
	}
	if name != "" {
		app.Name = &name
	}

	plan, err := cmd.Flags().GetString("plan")
	if err != nil {
		return app, err
	}
	if plan != "" {
		app.Plan = &plan
	}

	binID, err := cmd.Flags().GetInt64("binary")
	if err != nil {
		return app, err
	}
	if binID != 0 {
		app.Binary = &binID
	}

	disabled, err := cmd.Flags().GetBool("disabled")
	if err != nil {
		return app, err
	}
	if disabled {
		app.Status = newPointer(2)
	} else {
		app.Status = newPointer(1)
	}

	env, err := getMapParamP(cmd, "env")
	if err != nil {
		return app, err
	}
	app.Env = &env

	rspHeaders, err := getMapParamP(cmd, "rsp_headers")
	if err != nil {
		return app, err
	}
	app.RspHeaders = &rspHeaders

	return app, nil
}

func getMapParamP(cmd *cobra.Command, name string) (map[string]string, error) {
	ret := make(map[string]string)
	slice, err := cmd.Flags().GetStringSlice(name)
	if err != nil || slice == nil || len(slice) == 0 {
		return ret, err
	}
	for _, entry := range slice {
		// expect entry in format either key=value or key=
		bits := strings.SplitN(entry, "=", 2)
		if len(bits) != 2 {
			return nil, fmt.Errorf(`malformed key-value field "%s": %s`, name, entry)
		}
		ret[bits[0]] = bits[1]
	}

	return ret, nil
}

func appStatusToString(s int) string {
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

func outputMap(m *map[string]string, title string) {
	if m != nil && *m != nil && len(*m) > 0 {
		fmt.Println(title + ":")
		table := make([][]string, 0, len(*m))
		for k, v := range *m {
			table = append(table, []string{"\t" + k, v})
		}
		output.Table(table, output.FmtHuman)
	}
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
