package output

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type outputFormat string

const (
	FmtHuman       outputFormat = "human"
	FmtJSON        outputFormat = "json"
	outputOption                = "output"
	tblColumnSpace              = 2
)

var globalFormat outputFormat

// implement pflag.Value interface
func (f *outputFormat) String() string {
	return string(*f)
}

func (f *outputFormat) Set(v string) error {
	switch v {
	case string(FmtHuman), string(FmtJSON):
		*f = outputFormat(v)
	case "":
		*f = FmtHuman
	default:
		return fmt.Errorf(`must be one of "%s", "%s"`, FmtHuman, FmtJSON)
	}
	return nil
}

func (f *outputFormat) Type() string {
	return fmt.Sprintf(`("%s" | "%s")`, FmtHuman, FmtJSON)
}

func FormatOption(cmd *cobra.Command) {
	cmd.PersistentFlags().VarP(&globalFormat, outputOption, "o", `Output format (default "human")`)
}

func Format(cmd *cobra.Command) outputFormat {
	return globalFormat
}

func Table(lines [][]string) {
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
