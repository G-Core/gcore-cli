package subnet

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/human"
)

var (
	client *cloud.ClientWithResponses

	projectID     int
	regionID      int
	waitForResult bool
)

var reSubnetName = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z 0-9._\\-]{1,61}[a-zA-Z0-9._]$")

func validateSubnetName(name string) *errors.CliError {
	if reSubnetName.MatchString(name) {
		return nil
	}

	return &errors.CliError{
		Err: fmt.Errorf("network name doesn't match requirements"),
		// TODO: Maybe show user regex isn't the best idea, because not many people understand them
		Hint: fmt.Sprintf("Network name should match regex: '%s'", reSubnetName.String()),
		Code: 1, // TODO: need a convention about error codes
	}
}

type subnet struct {
	Id                     string
	Name                   string
	Cidr                   string
	IpVersion              cloud.SubnetSchemaIpVersion
	NetworkId              string
	TotalIps               int
	AvailableIps           int
	EnableDhcp             bool
	GatewayIp              *string
	HasRouter              bool
	ConnectToNetworkRouter bool
	HostRoutes             []cloud.NeutronRouteSchema
	DnsNameservers         []string
	Metadata               []cloud.MetadataItemSchema
	ProjectId              int
	Region                 string
	CreatedAt              string
	UpdatedAt              string
}

func toView(instance cloud.SubnetSchema) subnet {
	return subnet{
		Id:                     instance.Id,
		Name:                   instance.Name,
		Cidr:                   instance.Cidr,
		IpVersion:              instance.IpVersion,
		NetworkId:              instance.NetworkId,
		TotalIps:               instance.TotalIps,
		AvailableIps:           instance.AvailableIps,
		EnableDhcp:             instance.EnableDhcp,
		GatewayIp:              instance.GatewayIp,
		HasRouter:              instance.HasRouter,
		ConnectToNetworkRouter: instance.ConnectToNetworkRouter,
		HostRoutes:             instance.HostRoutes,
		DnsNameservers:         instance.DnsNameservers,
		Metadata:               instance.Metadata,
		ProjectId:              instance.ProjectId,
		Region:                 instance.Region,
		CreatedAt:              instance.CreatedAt,
		UpdatedAt:              instance.UpdatedAt,
	}
}

func init() {
	human.RegisterMarshalerFunc(cloud.SubnetSchema{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		instance := i.(cloud.SubnetSchema)
		s := toView(instance)

		return human.Marshal(s, nil)
	})

	human.RegisterMarshalerFunc([]cloud.SubnetSchema{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		instance := i.([]cloud.SubnetSchema)
		subnets := make([]subnet, len(instance))
		for i, s := range instance {
			subnets[i] = toView(s)
		}

		return human.Marshal(subnets, nil)
	})

	human.RegisterMarshalerFunc(cloud.SubnetListSchema{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		instance := i.(cloud.SubnetListSchema)
		subnets := make([]subnet, len(instance.Results))
		for i, s := range instance.Results {
			subnets[i] = toView(s)
		}

		return human.Marshal(subnets, nil)
	})
}

// top-level subnet command
func Commands() *cobra.Command {
	// subnetCmd represents the subnet command
	var subnetCmd = &cobra.Command{
		Use:   "subnet",
		Short: "Cloud subnet management commands",
		Long:  ``, // TODO:
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		GroupID: "cloud",
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			var (
				ctx = cmd.Context()
			)
			profile, err := core.GetClientProfile(ctx)
			if err != nil {
				return err
			}
			baseUrl := *profile.ApiUrl
			authFunc := core.ExtractAuthFunc(ctx)

			if profile.Local != nil && !*profile.Local {
				baseUrl += "/fastedge"
			}

			client, err = cloud.NewClientWithResponses(baseUrl, cloud.WithRequestEditorFn(authFunc))
			if err != nil {
				return fmt.Errorf("cannot init SDK: %w", err)
			}

			waitForResult = cmd.Flag("wait").Value.String() == "true"

			return nil
		},
	}

	subnetCmd.AddCommand(create(), show(), list(), update(), deleteCmd())
	return subnetCmd
}
