package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/G-core/cli/pkg/sdk"
	"github.com/spf13/cobra"
)

const (
	sourceStdin     = "-"
	wasmContentType = "application/octet-stream"
	tblColumnSpace  = 2
)

// top-level FastEdge command
func fastedge(client *sdk.ClientWithResponses) *cobra.Command {
	var cmdFastedge = &cobra.Command{
		Use:   "fastedge <subcommand>",
		Short: "Gcore Edge compute solution",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	cmdFastedge.AddCommand(
		app(client),
		binary(client),
	)

	return cmdFastedge
}

func binary(client *sdk.ClientWithResponses) *cobra.Command {
	var cmdBin = &cobra.Command{
		Use:   "binary <subcommand>",
		Short: "Binary-related commands",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdBinList = &cobra.Command{
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

			if outFormat(cmd) == outputJSON {
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
			outputTable(table)

			return nil
		},
	}

	var cmdBinAdd = &cobra.Command{
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
	cmdBinAdd.Flags().StringP("file", "f", sourceStdin, "Wasm binary filename (by default - stdin)")

	cmdBin.AddCommand(cmdBinList, cmdBinAdd)

	return cmdBin
}

// app-related commands
func app(client *sdk.ClientWithResponses) *cobra.Command {
	var cmdApp = &cobra.Command{
		Use:   "app <subcommand>",
		Short: "App-related commands",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdAppList = &cobra.Command{
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

			if outFormat(cmd) == outputJSON {
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
			outputTable(table)
			return nil
		},
	}

	var cmdAppGet = &cobra.Command{
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
			rsp, err := client.GetAppWithResponse(context.Background(), id)
			if err != nil {
				return fmt.Errorf("getting app detail: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting app details: %s", string(rsp.Body))
			}

			if outFormat(cmd) == outputJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf(
				"Name:\t%s\nBinary:\t%d\nPlan:\t%s\nStatus:\t%s\nUrl:\t%s\n",
				*(rsp.JSON200.Name),
				rsp.JSON200.Binary,
				rsp.JSON200.Plan,
				appStatusToString(rsp.JSON200.Status),
				*(rsp.JSON200.Url),
			)
			return nil
		},
	}

	cmdApp.AddCommand(cmdAppList, cmdAppGet)
	return cmdApp
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

func outputTable(lines [][]string) {
	// find max width for every column
	widths := make([]int, len(lines[0]))
	for _, line := range lines {
		for i, cell := range line {
			l := len(cell)
			if l > widths[i] {
				widths[i] = l
			}
		}
	}

	for _, line := range lines {
		for i, cell := range line {
			fmt.Print(cell + strings.Repeat(" ", widths[i]-len(cell)+tblColumnSpace))
		}
		fmt.Println()
	}
}

func unrefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
