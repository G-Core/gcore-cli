package region

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

type region struct {
	ID          int
	Name        string // DisplayName
	Ai          bool
	AiGpu       bool
	Baremetal   bool
	Vm          bool
	K8s         bool
	Kvm         bool
	Sfs         bool
	AccessLevel cloud.RegionSchemaAccessLevel
	Zone        cloud.RegionSchemaZone
	State       cloud.RegionSchemaState
	Country     cloud.RegionSchemaCountry
}

func toRegion(schema cloud.RegionSchema) region {
	return region{
		ID:          schema.Id,
		Name:        schema.DisplayName,
		Ai:          schema.HasAi,
		AiGpu:       schema.HasAiGpu,
		Baremetal:   schema.HasBaremetal,
		Vm:          schema.HasBasicVm,
		K8s:         schema.HasK8s,
		Kvm:         schema.HasKvm,
		Sfs:         schema.HasSfs,
		AccessLevel: schema.AccessLevel,
		Zone:        schema.Zone,
		State:       schema.State,
		Country:     schema.Country,
	}
}

func init() {
	human.RegisterMarshalerFunc(cloud.RegionSchema{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		schema := i.(cloud.RegionSchema)

		return human.Marshal(toRegion(schema), opt)
	})

	human.RegisterMarshalerFunc([]cloud.RegionSchema{}, func(i interface{}, opt *human.MarshalOpt) (string, error) {
		schemas := i.([]cloud.RegionSchema)

		list := make([]region, len(schemas))
		for idx, item := range schemas {
			list[idx] = toRegion(item)
		}

		return human.Marshal(list, opt)
	})
}

var client *cloud.ClientWithResponses

func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "region",
		Short:   "Cloud region commands",
		GroupID: "cloud",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			profile, err := core.GetClientProfile(ctx)
			if err != nil {
				return err
			}

			baseUrl := *profile.ApiUrl
			if !profile.IsLocal() {
				baseUrl += "/cloud"
			}

			authFunc := core.ExtractAuthFunc(ctx)
			client, err = cloud.NewClientWithResponses(baseUrl, cloud.WithRequestEditorFn(authFunc))
			if err != nil {
				return fmt.Errorf("cannot init SDK: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
			}

			return nil
		},
	}

	cmd.AddCommand(list())

	return cmd
}

// TODO: filter flag --filter=ai,edge - shows all regions with edge access and ai capability
func list() *cobra.Command {
	var showAll bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Displays regions available to user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			resp, err := client.GetRegionWithResponse(ctx, nil)
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

			var regions []cloud.RegionSchema
			for _, item := range resp.JSON200.Results {
				if showAll {
					regions = append(regions, item)

					continue
				}

				if item.State == cloud.RegionSchemaStateACTIVE {
					regions = append(regions, item)
				}
			}

			output.Print(regions)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "", false, "Show all regions (even not active)")

	return cmd
}
