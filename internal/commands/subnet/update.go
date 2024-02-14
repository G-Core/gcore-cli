package subnet

import (
	"net/http"

	"github.com/spf13/cobra"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
	"github.com/G-core/gcore-cli/internal/errors"
	"github.com/G-core/gcore-cli/internal/output"
	"github.com/G-core/gcore-cli/internal/sflags"
)

func update() *cobra.Command {
	var (
		name        string
		dhcp        = true
		dnsServers  []string
		sHostRoutes []string
		gateway     string
	)

	// createCmd represents the create command
	var cmd = &cobra.Command{
		Use:     "update <id> <flags>",
		Aliases: []string{"u"},
		Short:   "Update a specific subnet",
		Long:    ``,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			var opts cloud.PatchSubnetSchema
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

			if len(name) > 0 {
				opts.Name = name
			}

			if len(dnsServers) > 0 {
				opts.DnsNameservers = dnsServers
			}

			opts.EnableDhcp = dhcp

			if len(gateway) > 0 {
				opts.GatewayIp = &gateway
			}

			resp, err := client.PatchSubnetInstanceWithResponse(cmd.Context(), projectID, regionID, id, opts)

			if err != nil {
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				return errors.ParseCloudErr(resp.Body)
			}

			output.Print(resp.JSON200)

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "", "", "change subnet name")
	cmd.PersistentFlags().BoolVarP(&dhcp, "enable-dhcp", "", true, "default true")
	cmd.PersistentFlags().StringArrayVarP(&dnsServers, "dns", "", []string{}, "list of DNS servers")
	cmd.PersistentFlags().StringArrayVarP(&sHostRoutes, "host-route", "", []string{}, "list of host routes")
	cmd.PersistentFlags().StringVarP(&gateway, "gateway", "", "", "gateway")

	return cmd
}
