package config

import (
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/output"
)

type profileView struct {
	Name   string
	ApiUrl *string
	ApiKey *string
}

func toProfileView(name string, profile *config.Profile) profileView {
	var pv = profileView{
		Name: name,
	}

	if profile.ApiUrl != nil {
		pv.ApiUrl = profile.ApiUrl
	}

	if profile.ApiKey != nil {
		pv.ApiKey = pointer.To(secureKey(profile.ApiKey))
	}

	return pv
}

func info() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Get information about config profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			profile, err := core.GetClientProfile(ctx)
			if err != nil {
				return err
			}

			output.Print(toProfileView(core.ExtractProfile(ctx), profile))

			return nil
		},
	}

	return cmd
}

func secureKey(key *string) string {
	if key == nil || *key == "" {
		return ""
	}

	var p1 = 0 + 5
	var p2 = len(*key) - 1 - 5
	if p1 > p2 {
		return "XXXXXX"
	}

	return strings.Join([]string{(*key)[0:p1], "XXXXXX", (*key)[p2 : len((*key))-1]}, "")
}
