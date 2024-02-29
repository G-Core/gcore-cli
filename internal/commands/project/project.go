package project

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/human"
	"github.com/G-core/gcore-cli/internal/output"
)

type project struct {
	Id          int
	Name        string
	Description *string
	State       string
	IsDefault   bool
	ClientId    int
	CreatedAt   string
	DeletedAt   *string
	TaskId      *string
}

func toProject(schema cloud.ProjectSerializer) project {
	return project{
		Id:          schema.Id,
		Name:        schema.Name,
		Description: schema.Description,
		State:       schema.State,
		IsDefault:   schema.IsDefault,
		ClientId:    schema.ClientId,
		CreatedAt:   schema.CreatedAt,
		DeletedAt:   schema.DeletedAt,
		TaskId:      schema.TaskId,
	}
}

func init() {
	human.RegisterMarshalerFunc(cloud.ProjectSerializer{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		schema := i.(cloud.ProjectSerializer)

		return human.Marshal(toProject(schema), opt)
	})

	human.RegisterMarshalerFunc([]cloud.ProjectSerializer{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		schemas := i.([]cloud.ProjectSerializer)

		list := make([]project, len(schemas))
		for idx, schema := range schemas {
			list[idx] = toProject(schema)
		}

		return human.Marshal(list, opt)
	})
}

var client *cloud.ClientWithResponses

func Commands() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "project",
		Short: "Commands to manage Cloud projects",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			profile, err := core.GetClientProfile(ctx)
			if err != nil {
				return err
			}

			baseUrl := *profile.ApiUrl
			if profile.Local != nil && !*profile.Local {
				baseUrl += "/cloud"
			}

			authFunc := core.ExtractAuthFunc(ctx)
			client, err = cloud.NewClientWithResponses(baseUrl, cloud.WithRequestEditorFn(authFunc))
			if err != nil {
				return fmt.Errorf("cannot init SDK: %w", err)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}

	cmd.AddCommand(list())

	return cmd
}

func list() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Displays users projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			resp, err := client.GetProjectListWithResponse(ctx, nil)
			if err != nil {
				return err
			}

			switch resp.StatusCode() {
			case http.StatusOK:
			case http.StatusNotFound:
				return &errors.CliError{
					Err:  fmt.Errorf("404 not found"),
					Hint: fmt.Sprintf("Check profile '%s' configuration: api-url and local", core.ExtractProfile(ctx)),
				}
			default:
				return errors.ParseCloudErr(resp.Body)
			}

			// TODO: process cases where client get redirected to api.gcore.com/docs
			if resp.JSON200 == nil {
				return &errors.CliError{
					Err:  fmt.Errorf("404 not found"),
					Hint: fmt.Sprintf("Check profile '%s' configuration: api-url and local", core.ExtractProfile(ctx)),
				}
			}

			output.Print(resp.JSON200.Results)

			return nil
		},
	}

	return cmd
}
