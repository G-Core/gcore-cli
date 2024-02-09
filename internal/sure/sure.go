package sure

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func AreYou(cmd *cobra.Command, message string) bool {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		fmt.Fprintf(os.Stderr, `"force" flag: %v\n`, err)
		return false
	}
	if force {
		return true
	}
	for {
		fmt.Printf("Are you sure to %s? [y/N] ", message)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(response)
		switch response {
		case "y", "yes", "yep", "yeah":
			return true
		case "", "n", "no", "nope":
			return false
		default:
			continue
		}
	}
}
