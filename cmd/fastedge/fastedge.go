package fastedge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/G-core/cli/pkg/output"
	"github.com/G-core/cli/pkg/sdk"
)

const (
	sourceStdin     = "-"
	wasmContentType = "application/octet-stream"
)

var client *sdk.ClientWithResponses

// top-level FastEdge command
func Commands(baseUrl string, authFunc func(ctx context.Context, req *http.Request) error) (*cobra.Command, error) {
	var cmdFastedge = &cobra.Command{
		Use:   "fastedge <subcommand>",
		Short: "Gcore Edge compute solution",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			client, err = sdk.NewClientWithResponses(
				baseUrl+"/fastedge",
				sdk.WithRequestEditorFn(authFunc),
			)
			if err != nil {
				return fmt.Errorf("cannot init SDK: %w", err)
			}
			return nil
		},
	}

	cmdFastedge.AddCommand(app(), binary(), plan())
	return cmdFastedge, nil
}

func binary() *cobra.Command {
	var cmdBin = &cobra.Command{
		Use:   "binary <subcommand>",
		Short: "Binary-related commands",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of client's binaries",
		Long:    ``,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.ListBinariesWithResponse(context.Background())
			if err != nil {
				return fmt.Errorf("getting the list of binaries: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the list of binaries: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if len(*rsp.JSON200) == 0 {
				fmt.Printf("you have no binaries\n")
				return nil
			}

			table := make([][]string, len(*rsp.JSON200)+1)
			table[0] = []string{"ID", "Status", "Unreferenced since"}
			for i, bin := range *rsp.JSON200 {
				table[i+1] = []string{
					strconv.FormatInt(bin.Id, 10),
					binStatusToString(bin.Status),
					unrefString(bin.UnrefSince),
				}
			}
			output.Table(table)

			return nil
		},
	}

	var cmdUpload = &cobra.Command{
		Use:     "add",
		Aliases: []string{"upload"},
		Short:   "Add new binary",
		Long:    ``,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := cmd.Flags().GetString("file")
			if err != nil {
				return errors.New("please specify binary filename")
			}
			r := os.Stdin
			if src != sourceStdin {
				r, err = os.Open(src)
				if err != nil {
					return fmt.Errorf("cannot open %s: %w", src, err)
				}
				defer r.Close()
			}

			rsp, err := client.StoreBinaryWithBodyWithResponse(
				context.Background(),
				wasmContentType,
				r,
			)
			if err != nil {
				return fmt.Errorf("cannot upload the binary: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("cannot upload the binary: %s", string(rsp.Body))
			}

			fmt.Printf("Uploaded binary with ID %d\n", *rsp.JSON200)

			return nil
		},
	}
	cmdUpload.Flags().StringP("file", "f", sourceStdin, "Wasm binary filename (by default - stdin)")

	cmdBin.AddCommand(cmdList, cmdUpload)

	return cmdBin
}

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
			if app.Binary == nil {
				return errors.New("binary must be specified")
			}
			if app.Plan == nil {
				return errors.New("plan must be specified")
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
	cmd.Flags().String("file", "", "Wasm binary filename")
	cmd.Flags().String("plan", "", "Plan name")
	cmd.Flags().Int("status", 1, "Status (0 - draft, 1 - enabled, 2 - disabled)")
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
	status, err := cmd.Flags().GetInt("status")
	if err != nil {
		return app, err
	}
	if status != -1 {
		app.Status = &status
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

func plan() *cobra.Command {
	var cmdPlan = &cobra.Command{
		Use:   "plan <subcommand>",
		Short: "Plan-related commands",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of available plans",
		Long:    ``,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.ListPlansWithResponse(context.Background())
			if err != nil {
				return fmt.Errorf("getting the list of plans: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the list of plans: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if len(*rsp.JSON200) == 0 {
				fmt.Printf("there are no plans\n")
				return nil
			}

			fmt.Println("Plan name")
			for _, plan := range *rsp.JSON200 {
				fmt.Println(plan)
			}

			return nil
		},
	}

	var cmdGet = &cobra.Command{
		Use:     "show <plan_name>",
		Aliases: []string{"get"},
		Short:   "Show plan details",
		Long:    ``,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.GetPlanWithResponse(
				context.Background(),
				args[0],
			)
			if err != nil {
				return fmt.Errorf("cannot get plan details: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("cannot get plan details: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf(
				"Memory limit:\t%s\nMax duration:\t%s\nMax requests:\t%d\n",
				units.HumanSize(float64(rsp.JSON200.MemLimit)),
				time.Duration(rsp.JSON200.MaxDuration)*time.Millisecond,
				rsp.JSON200.MaxSubrequests,
			)

			return nil
		},
	}

	cmdPlan.AddCommand(cmdList, cmdGet)

	return cmdPlan
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

func binStatusToString(s int) string {
	switch s {
	case 0:
		return "pending"
	case 1:
		return "compiled"
	case 2:
		return "compilation failed (see errors)"
	case 3:
		return "compilation failed (no errors reported)"
	case 4:
		return "compilatoin result exceeds the size limit"
	case 5:
		return "unsupported source language"
	}
	return "unknown"

}

func unrefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
