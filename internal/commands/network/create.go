package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
)

// ^[a-zA-Z0-9][a-zA-Z 0-9._\-]{1,61}[a-zA-Z0-9._]$
func validateNetworkName(name string) *errors.CliError {
	if reNetworkName.MatchString(name) {
		return nil
	}

	return &errors.CliError{
		Err: fmt.Errorf("network name doesn't match requirements"),
		// TODO: Maybe show user regex isn't the best idea, because not many people understand them
		Hint: fmt.Sprintf("Network name should match regex: '%s'", reNetworkName.String()),
		Code: 1, // TODO: need a convention about error codes
	}
}

func validateNetworkType(netType string) *errors.CliError {
	netType = strings.ToLower(netType)
	switch netType {
	case string(cloud.CreateNetworkSchemaTypeVxlan):
		return nil
	case string(cloud.CreateNetworkSchemaTypeVlan):
		return nil
	}

	return &errors.CliError{
		Err:  fmt.Errorf("wrong network type"),
		Hint: "Only vlan or vxlan network type is allowed.",
		Code: 1,
	}
}

// TODO: Metadata flag
var (
	name        string
	networkType string
	router      bool
)

func create() *cobra.Command {
	// createCmd represents the create command
	var createCmd = &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"c"},
		Short:   "Create a network",
		Long:    ``,
		Args:    cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx := cmd.Context()
			projectID, err = core.ExtractCloudProject(ctx)
			if err != nil {
				return err
			}

			regionID, err = core.ExtractCloudRegion(ctx)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Autogenerate names for networks as option
			name = args[0]

			// TODO: Refactor validation
			if err := validateNetworkName(name); err != nil {
				return err
			}

			if err := validateNetworkType(networkType); err != nil {
				return err
			}

			resp, err := client.PostNetworkWithResponse(cmd.Context(), projectID, regionID, cloud.CreateNetworkSchema{
				Name:         name,
				Type:         cloud.CreateNetworkSchemaType(networkType),
				CreateRouter: router,
			})

			if err != nil {
				return err
			}

			if !waitForResult {
				return nil
			}

			var networkID string
			_, err = cloud.WaitTaskAndReturnResult(cmd.Context(), client, resp.JSON200.Tasks[0], true, time.Second*5, func(task *cloud.TaskSchema) (any, error) {
				networkID = task.CreatedResources.Networks[0]
				return nil, nil
			})

			if err != nil {
				return &errors.CliError{
					Err: fmt.Errorf("task %s: %w", resp.JSON200.Tasks[0], err),
				}
			}

			return displayNetwork(cmd.Context(), networkID)
		},
	}

	createCmd.PersistentFlags().StringVarP(&networkType, "type", "",
		string(cloud.CreateNetworkSchemaTypeVxlan),
		"Network type, vlan or vxlan network type is allowed. Default value is vxlan")
	createCmd.PersistentFlags().BoolVarP(&router, "router", "", true, "Create router. Default is true.")

	return createCmd
}
