package fastedge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/G-core/cli/pkg/output"
)

const (
	sourceStdin     = "-"
	wasmContentType = "application/octet-stream"
)

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

			id, err := uploadBinary(src)
			if err != nil {
				return err
			}

			fmt.Printf("Uploaded binary with ID %d\n", id)

			return nil
		},
	}
	cmdUpload.Flags().StringP("file", "f", sourceStdin, "Wasm binary filename ('-' means stdin)")

	cmdBin.AddCommand(cmdList, cmdUpload)

	return cmdBin
}

func uploadBinary(src string) (int64, error) {
	r := os.Stdin
	var err error
	if src != sourceStdin {
		r, err = os.Open(src)
		if err != nil {
			return 0, fmt.Errorf("cannot open %s: %w", src, err)
		}
		defer r.Close()
	}

	rsp, err := client.StoreBinaryWithBodyWithResponse(
		context.Background(),
		wasmContentType,
		r,
	)
	if err != nil {
		return 0, fmt.Errorf("cannot upload the binary: %w", err)
	}
	if rsp.StatusCode() != http.StatusOK {
		return 0, fmt.Errorf("cannot upload the binary: %s", string(rsp.Body))
	}

	return *rsp.JSON200, nil
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
