package fastedge

import (
	"fmt"

	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cobra"

	sdk "github.com/G-Core/FastEdge-client-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
)

var client *sdk.ClientWithResponses

// top-level FastEdge command
func Commands() *cobra.Command {
	var cmdFastedge = &cobra.Command{
		Use:     "fastedge <subcommand>",
		Short:   "Gcore Edge compute solution",
		Long:    ``,
		GroupID: "fastedge",
		Args:    cobra.MinimumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var (
				err error
				ctx = cmd.Context()
			)
			profile, err := core.GetClientProfile(ctx)
			if err != nil {
				return err
			}
			url := *profile.ApiUrl
			authFunc := core.ExtractAuthFunc(ctx)

			if profile.Local != nil && !*profile.Local {
				url += "/fastedge"
			}

			client, err = sdk.NewClientWithResponses(
				url,
				sdk.WithRequestEditorFn(authFunc),
			)
			if err != nil {
				return fmt.Errorf("cannot init SDK: %w", err)
			}

			carbon.SetDefault(carbon.Default{
				Timezone: carbon.UTC,
				Locale:   "en",
			})

			return nil
		},
	}

	cmdFastedge.AddCommand(app(), binary(), plan(), stat(), logs())
	return cmdFastedge
}

func newPointer[T any](val T) *T {
	return &val
}
