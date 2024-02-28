package config

import (
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/config"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/output"
)

func getProfileField(profile *config.Profile, key string) (reflect.Value, error) {
	field := reflect.ValueOf(profile).Elem().FieldByName(strcase.ToCamel(key))
	reflect.ValueOf(profile).Elem().FieldByNameFunc(func(s string) bool {
		return key == strcase.ToKebab(s)
	})

	if !field.IsValid() {
		return reflect.ValueOf(nil), fmt.Errorf("invalid key: %s", key)
	}

	return field, nil
}

func getProfileValue(profile *config.Profile, fieldName string) (interface{}, error) {
	field, err := getProfileField(profile, fieldName)
	if err != nil {
		return nil, err
	}
	return field.Interface(), nil
}

func get() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get <property>",
		Short: "Get property value from the config file",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}

			ctx := cmd.Context()
			profileName := core.ExtractProfile(ctx)
			cfg := core.ExtractConfig(ctx)

			profile, err := cfg.GetProfile(profileName)
			if err != nil {
				return err
			}

			value, err := getProfileValue(profile, args[0])
			if err != nil {
				return err
			}

			output.Print(value)

			return nil
		},
	}

	return cmd
}
