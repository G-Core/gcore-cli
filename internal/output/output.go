package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/G-core/gcore-cli/internal/human"

	"github.com/spf13/cobra"
)

type outputFormat string

const (
	FmtHuman       outputFormat = "human"
	FmtJSON        outputFormat = "json"
	FmtCSV         outputFormat = "csv"
	outputOption                = "output"
	tblColumnSpace              = 2
	csvDelimiter                = ","
)

var globalFormat = FmtHuman

// implement pflag.Value interface
func (f *outputFormat) String() string {
	return string(*f)
}

func (f *outputFormat) Set(v string) error {
	switch v {
	case string(FmtHuman), string(FmtJSON), string(FmtCSV):
		*f = outputFormat(v)
	case "":
		*f = FmtHuman
	default:
		return fmt.Errorf(`must be one of "%s", "%s", "%s"`, FmtHuman, FmtJSON, FmtCSV)
	}
	return nil
}

func (f *outputFormat) Type() string {
	return fmt.Sprintf(`("%s" | "%s" | "%s")`, FmtHuman, FmtJSON, FmtCSV)
}

func FormatOption(cmd *cobra.Command) {
	cmd.PersistentFlags().VarP(&globalFormat, outputOption, "o", `Output format ("json", "csv" or "human', default "human")`)
}

func Format(cmd *cobra.Command) outputFormat {
	return globalFormat
}

func IsJSON() bool {
	return globalFormat == FmtJSON
}

func Print(data any) {
	var (
		body string
		err  error
	)

	switch globalFormat {
	case FmtJSON:
		bytes, err := json.Marshal(data)
		if err == nil {
			body = string(bytes)
		}
	case FmtHuman:
		body, err = human.Marshal(data, nil)
		// TODO: Support CSV?
	default:
		err = fmt.Errorf("format '%s' is not supported", globalFormat)
	}

	if err != nil {
		body, _ = human.Marshal(err, nil)
	}

	fmt.Println(body)
}

func Table(lines [][]string, format outputFormat) {
	if format == FmtCSV {
		w := csv.NewWriter(os.Stdout)
		w.WriteAll(lines)
		return
	}

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
