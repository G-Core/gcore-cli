package network

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/cli/pkg/human"

	"github.com/spf13/cobra"
)

var (
	client *cloud.ClientWithResponses

	projectID int
	regionID  int
)

type network struct {
	// Id Network ID
	Id string
	// Name Network name
	Name string
	// Type Network type (vlan, vxlan)
	Type string
	// External True if the network has `router:external` attribute
	External bool
	// Default True if the network has is_default attribute
	Default bool
	// Shared True when the network is shared with your project by external owner
	Shared bool
	// Mtu MTU (maximum transmission unit). Default value is 1450
	Mtu int
	// Subnets List of subnetworks
	Subnets []string
	// Metadata Network metadata
	Metadata []cloud.MetadataItemSchema
	// SegmentationId Id of network segment
	SegmentationId int
	// ProjectId Project ID
	ProjectId int
	// Region Region name
	Region string
	// RegionId Region ID
	RegionId int
	// CreatedAt Datetime when the network was created
	CreatedAt string
	// UpdatedAt Datetime when the network was last updated
	UpdatedAt string
}

func init() {
	human.RegisterMarshalerFunc(cloud.NetworkSchema{}, func(i interface{}, opt *human.MarshalOpt) (body string, err error) {
		instance := i.(cloud.NetworkSchema)
		s := network{
			Id:             instance.Id,
			Name:           instance.Name,
			Type:           instance.Type,
			External:       instance.External,
			Default:        instance.Default,
			Shared:         instance.Shared,
			Mtu:            instance.Mtu,
			Subnets:        instance.Subnets,
			Metadata:       instance.Metadata,
			SegmentationId: instance.SegmentationId,
			ProjectId:      instance.ProjectId,
			Region:         instance.Region,
			RegionId:       instance.RegionId,
			CreatedAt:      instance.CreatedAt,
			UpdatedAt:      instance.UpdatedAt,
		}

		return human.Marshal(s, nil)
	})

	human.RegisterMarshalerFunc([]cloud.NetworkSchema{}, func(i interface{}, opt *human.MarshalOpt) (body string, err error) {
		instances := i.([]cloud.NetworkSchema)
		s := make([]network, len(instances))
		for i, instance := range instances {
			s[i] = network{
				Id:             instance.Id,
				Name:           instance.Name,
				Type:           instance.Type,
				External:       instance.External,
				Default:        instance.Default,
				Shared:         instance.Shared,
				Mtu:            instance.Mtu,
				Subnets:        instance.Subnets,
				Metadata:       instance.Metadata,
				SegmentationId: instance.SegmentationId,
				ProjectId:      instance.ProjectId,
				Region:         instance.Region,
				RegionId:       instance.RegionId,
				CreatedAt:      instance.CreatedAt,
				UpdatedAt:      instance.UpdatedAt,
			}
		}

		return human.Marshal(s, nil)
	})
}

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
