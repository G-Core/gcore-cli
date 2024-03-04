package subnet

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/core"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/sflags"
)

func create() *cobra.Command {
	var (
		networkID   string
		cidr        string
		dhcp        = true
		ipVersion   = 4
		router      = false
		dnsServers  []string
		sHostRoutes []string
		gateway     string
	)

	// createCmd represents the create command
	var cmd = &cobra.Command{
		Use:     "create <name> <flags>",
		Aliases: []string{"c"},
		Short:   "Create a subnet for specific network",
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
			var opts cloud.CreateSubnetSchema

			opts.Name = args[0]
			if err := validateSubnetName(opts.Name); err != nil {
				return err
			}

			// TODO: Validation
			if networkID == "" || cidr == "" {
				return &errors.CliError{
					Message: "One of required flags is missing",
					// TODO: Possibility to automate this
					Hint: "--cidr or --network are required for this command",
				}
			}
			opts.NetworkId = networkID
			opts.Cidr = cidr

			if len(sHostRoutes) > 0 {
				sm, err := sflags.ParseSlice(sHostRoutes)
				if err != nil {
					return err
				}

				for _, m := range sm {
					if m["destination"] == "" || m["nexthop"] == "" {
						return &errors.CliError{
							Message: "One of host route fields is missing",
							Hint:    "To add host route use --host-route destination:value1,nexthop:value2",
						}
					}

					opts.HostRoutes = append(opts.HostRoutes, cloud.NeutronRouteSchema{
						Destination: m["destination"],
						Nexthop:     m["nexthop"],
					})
				}
			}

			if gateway != "" {
				opts.GatewayIp = &gateway
			}
			opts.ConnectToNetworkRouter = router
			opts.EnableDhcp = dhcp
			opts.IpVersion = cloud.CreateSubnetSchemaIpVersion(ipVersion)

			if len(dnsServers) > 0 {
				opts.DnsNameservers = dnsServers
			}

			resp, err := client.PostSubnetWithResponse(cmd.Context(), projectID, regionID, opts)
			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.ParseCloudErr(resp.Body)
			}

			if !waitForResult {
				return nil
			}

			var subnetID string
			_, err = cloud.WaitTaskAndReturnResult(cmd.Context(), client, resp.JSON200.Tasks[0], true, time.Second*5, func(task *cloud.TaskSchema) (any, error) {
				subnetID = task.CreatedResources.Subnets[0]
				return nil, nil
			})

			if err != nil {
				return &errors.CliError{
					Err: fmt.Errorf("task %s: %w", resp.JSON200.Tasks[0], err),
				}
			}

			return displaySubnet(cmd.Context(), subnetID)
		},
	}

	cmd.PersistentFlags().StringVarP(&networkID, "network", "", "", "network id")
	cmd.RegisterFlagCompletionFunc("network", core.NetworkCompletion)
	cmd.PersistentFlags().StringVarP(&cidr, "cidr", "", "", "subnet CIDR")
	cmd.PersistentFlags().BoolVarP(&dhcp, "enable-dhcp", "", true, "default true")
	cmd.PersistentFlags().IntVar(&ipVersion, "ip-version", 4, "IP version. default 4")
	cmd.RegisterFlagCompletionFunc("ip-version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"4\tIPv4", "6\tIPv6"}, cobra.ShellCompDirectiveDefault
	})
	cmd.PersistentFlags().BoolVarP(&router, "enable-router", "", false, "")
	cmd.PersistentFlags().StringArrayVarP(&dnsServers, "dns", "", []string{}, "list of DNS servers")
	cmd.PersistentFlags().StringArrayVarP(&sHostRoutes, "host-route", "", []string{}, "list of host routes")
	cmd.PersistentFlags().StringVarP(&gateway, "gateway", "", "", "gateway IP")

	return cmd
}
