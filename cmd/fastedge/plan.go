package fastedge

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/G-core/cli/pkg/output"
)

func plan() *cobra.Command {
	var cmdPlan = &cobra.Command{
		Use:   "plan <subcommand>",
		Short: "Plan-related commands",
		Long:  `Plan is a set of limits for the application. Different plans may imply different rates,
so chose the plan to minimise costs. However, using plan smaller than needed
may result in excessive timeouts and out-of-memory errors.`,
		Args:  cobra.MinimumNArgs(1),
	}

	var cmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show list of available plans",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.ListPlansWithResponse(context.Background())
			if err != nil {
				return fmt.Errorf("getting the list of plans: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the list of plans: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if len(*rsp.JSON200) == 0 {
				fmt.Printf("there are no plans\n")
				return nil
			}

			fmt.Println("Plan name")
			for _, plan := range *rsp.JSON200 {
				fmt.Println(plan)
			}

			return nil
		},
	}

	var cmdGet = &cobra.Command{
		Use:     "show <plan_name>",
		Aliases: []string{"get"},
		Short:   "Show plan details",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.GetPlanWithResponse(
				context.Background(),
				args[0],
			)
			if err != nil {
				return fmt.Errorf("cannot get plan details: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("cannot get plan details: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf(
				"Memory limit:\t%s\nMax duration:\t%s\nMax requests:\t%d\n",
				units.HumanSize(float64(rsp.JSON200.MemLimit)),
				time.Duration(rsp.JSON200.MaxDuration)*time.Millisecond,
				rsp.JSON200.MaxSubrequests,
			)

			return nil
		},
	}

	cmdPlan.AddCommand(cmdList, cmdGet)

	return cmdPlan
}
