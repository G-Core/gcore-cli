package network

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/spf13/cobra"
)

var (
	client *cloud.ClientWithResponses

	projectID int
	regionID  int
)

// top-level cloud network command
func Commands(baseUrl string, authFunc func(ctx context.Context, req *http.Request) error) (*cobra.Command, error) {
	// networkCmd represents the network command
	var networkCmd = &cobra.Command{
		Use:   "network",
		Short: "Cloud network management commands",
		Long:  ``, // TODO:
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			client, err = cloud.NewClientWithResponses(baseUrl, cloud.WithRequestEditorFn(authFunc))
			if err != nil {
				return fmt.Errorf("cannot init SDK: %w", err)
			}

			fProject := cmd.Flag("project")
			if fProject == nil {
				return fmt.Errorf("can't find --project flag")
			}

			projectID, err = strconv.Atoi(fProject.Value.String())
			if err != nil {
				return fmt.Errorf("--project flag value must to be int: %w", err)
			}

			fRegion := cmd.Flag("region")
			if fRegion == nil {
				return fmt.Errorf("can't find --region flag")
			}

			regionID, err = strconv.Atoi(fRegion.Value.String())
			if err != nil {
				return fmt.Errorf("--region flag value must to be int: %w", err)
			}

			return nil
		},
	}

	networkCmd.AddCommand(create(), show(), list(), update(), delete())
	return networkCmd, nil
}
