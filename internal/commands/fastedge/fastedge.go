package fastedge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cobra"

	sdk "github.com/G-Core/FastEdge-client-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
)

const (
	versionHeaderName = "Fastedge-Sdk-Version"
	SDKpackage        = "github.com/G-Core/FastEdge-client-sdk-go"
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

			if !profile.IsLocal() {
				url += "/fastedge"
			}

			client, err = sdk.NewClientWithResponses(
				url,
				sdk.WithRequestEditorFn(authFunc),
				sdk.WithRequestEditorFn(addSDKversionHeader),
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

	cmdFastedge.AddCommand(app(), binary(), stat(), logs())
	return cmdFastedge
}

func newPointer[T any](val T) *T {
	return &val
}

func addSDKversionHeader(ctx context.Context, req *http.Request) error {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range bi.Deps {
			if dep.Path == SDKpackage {
				ver := strings.SplitN(dep.Version, "-", 2) // drop revision info
				req.Header.Set(versionHeaderName, ver[0])
				return nil
			}
		}
	}
	return nil
}

type errResponse struct {
	Error string `json:"error"`
}

func extractErrorMessage(rspBuf []byte) string {
	var rsp errResponse
	if err := json.Unmarshal(rspBuf, &rsp); err == nil {
		return rsp.Error
	}
	return string(rspBuf)
}
