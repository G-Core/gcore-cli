package fastedge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/G-core/cli/pkg/output"
	"github.com/G-core/cli/pkg/sdk"
)

// app-related commands
func app() *cobra.Command {
	var cmdApp = &cobra.Command{
		Use:   "app <subcommand>",
		Short: "App-related commands",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"add"},
		Short:   "Add new app",
		Long:    ``,
		Args:    cobra.NoArgs,
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

	var cmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of client's apps",
		Long:    ``,
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
			output.Table(table)
			return nil
		},
	}

	var cmdGet = &cobra.Command{
		Use:     "show <app_id>",
		Aliases: []string{"get"},
		Short:   "Show app details",
		Long:    ``,
		Args:    cobra.ExactArgs(1),
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
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			status := 1
			rsp, err := client.PatchAppWithResponse(
				context.Background(),
				id,
				sdk.App{Status: &status},
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
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("parsing app id: %w", err)
			}
			status := 2
			rsp, err := client.PatchAppWithResponse(
				context.Background(),
				id,
				sdk.App{Status: &status},
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

	cmdApp.AddCommand(
		cmdList,
		cmdGet,
		cmdEnable,
		cmdDisable,
		cmdCreate,
	)
	return cmdApp
}

func appPropertiesFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "App name")
	cmd.Flags().Int64("binary", -1, "Wasm binary id")
	cmd.Flags().StringP("file", "f", sourceStdin, "Wasm binary filename ('-' means stdin)")
	cmd.Flags().String("plan", "", "Plan name")
	cmd.Flags().Bool("disabled", false, "Set status to 'disabled'")
	cmd.Flags().StringSlice("env", nil, "Environment, in name=value format")
	cmd.Flags().StringSlice("rsp_headers", nil, "Response headers to add, in name=value format")
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
	if binID != -1 {
		app.Binary = &binID
	}

	status := 1
	disabled, err := cmd.Flags().GetBool("disabled")
	if err != nil {
		return app, err
	}
	if disabled {
		status = 2
	}
	app.Status = &status

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
		output.Table(table)
	}
}
