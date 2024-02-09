package subnet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// top-level subnet command
func Commands(baseUrl string, authFunc func(ctx context.Context, req *http.Request) error) (*cobra.Command, error) {
	// subnetCmd represents the subnet command
	var subnetCmd = &cobra.Command{
		Use:   "subnet",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("subnet called")
		},
	}

	return subnetCmd, nil
}
