package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type outputFormat string

const (
	outputHuman  outputFormat = "human"
	outputJSON   outputFormat = "json"
	outputOption              = "output"
)

var globalFormat outputFormat

// implement pflag.Value interface
func (f *outputFormat) String() string {
	return string(*f)
}

func (f *outputFormat) Set(v string) error {
	switch v {
	case string(outputHuman), string(outputJSON):
		*f = outputFormat(v)
	case "":
		*f = outputHuman
	default:
		return fmt.Errorf(`must be one of "%s" "%s"`, string(outputHuman), string(outputJSON))
	}
	return nil
}

func (f *outputFormat) Type() string {
	return "outputFormat"
}

func formatOption(cmd *cobra.Command) {
	cmd.PersistentFlags().VarP(&globalFormat, outputOption, "o", "Output format ('json' or 'human', default - 'human')")
}

func outFormat(cmd *cobra.Command) outputFormat {
	return globalFormat
}
