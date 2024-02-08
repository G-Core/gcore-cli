package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/cli/pkg/errors"
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

var (
	name        string
	networkType string
	router      bool
)

func create() *cobra.Command {
	// createCmd represents the create command
	var createCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a network",
		Long:    ``,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.PersistentFlags().StringVar(&networkType, "type",
				string(cloud.CreateNetworkSchemaTypeVxlan),
				"Network type, vlan or vxlan network type is allowed. Default value is vxlan")
			cmd.PersistentFlags().BoolVar(&router, "router", true, "Create router. Default is true.")

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

			_, err := client.PostNetworkWithResponse(cmd.Context(), projectID, regionID, cloud.CreateNetworkSchema{
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

			return nil
		},
	}

	return createCmd
}
