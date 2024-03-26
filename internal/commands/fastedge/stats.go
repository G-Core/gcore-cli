package fastedge

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	sdk "github.com/G-Core/FastEdge-client-sdk-go"
	"github.com/golang-module/carbon/v2"
	"github.com/spf13/cobra"

	"github.com/G-core/gcore-cli/internal/output"
)

func stat() *cobra.Command {
	var cmdStat = &cobra.Command{
		Use:   "stats <subcommand>",
		Short: "Statistics",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			rsp, err := client.GetClientMeWithResponse(context.Background())
			if err != nil {
				return fmt.Errorf("getting the statistics: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("getting the statistics: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			fmt.Printf("Status:\t\t%s\nApps:\t\t%d",
				appStatusToString(rsp.JSON200.Status),
				rsp.JSON200.AppCount,
			)
			if rsp.JSON200.AppLimit > 0 {
				fmt.Printf(" out of allowed %d", rsp.JSON200.AppLimit)
			}
			fmt.Printf("\nHourly calls:\t%d", rsp.JSON200.HourlyConsumption)
			if rsp.JSON200.HourlyLimit > 0 {
				fmt.Printf(" out of allowed %d", rsp.JSON200.HourlyLimit)
			}
			fmt.Printf("\nDaily calls:\t%d", rsp.JSON200.DailyConsumption)
			if rsp.JSON200.DailyLimit > 0 {
				fmt.Printf(" out of allowed %d", rsp.JSON200.DailyLimit)
			}
			fmt.Println()

			return nil
		},
	}

	var cmdCalls = &cobra.Command{
		Use:     "calls <app_name>",
		Aliases: []string{"calls"},
		Short:   "Show app calls statistic",
		Long: `Show number of app calls, grouped by time slots and HTTP statuses.
By default it reports every hour from the beginning of current day (UTC),
but you can change reporting interval using "--from" and "--to" flags
(specifying date/time in format "YYYY-MM-DD HH:mm:SS", where either date or time,
can be omitted, or as UNIX timestamp) and reporting step duration with flag
"--step" (in seconds).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getAppIdByName(args[0])
			if err != nil {
				return fmt.Errorf("cannot find app by name: %w", err)
			}

			from, err := parseTimeFlag(cmd, "from")
			if err != nil {
				return fmt.Errorf("cannot parse 'from' time: %w", err)
			}

			to, err := parseTimeFlag(cmd, "to")
			if err != nil {
				return fmt.Errorf("cannot parse 'to' time: %w", err)
			}

			step, err := cmd.Flags().GetInt("step")
			if err != nil {
				return fmt.Errorf("cannot parse reporting step: %w", err)
			}

			rsp, err := client.AppCallsWithResponse(
				context.Background(),
				id,
				&sdk.AppCallsParams{
					From: from,
					To:   to,
					Step: step,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot get statistics: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("cannot get statistics: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if len(rsp.JSON200.Stats) == 0 {
				fmt.Println("No data to report")
				return nil
			}

			// we don't know which statuses we see, so collect the info about statuses
			// and make sparse matrix for counts by status
			statusCols := make(map[int]int)
			counts := make([][]int, len(rsp.JSON200.Stats))
			for i, slot := range rsp.JSON200.Stats {
				line := make([]int, len(statusCols))
				for _, count := range slot.CountByStatus {
					col, ok := statusCols[count.Status]
					if ok {
						line[col] = count.Count
					} else {
						statusCols[count.Status] = len(statusCols)
						line = append(line, count.Count)
					}
				}
				counts[i] = line
			}

			// determine correct column order
			statuses := make([]int, 0, len(statusCols))
			for k := range statusCols {
				statuses = append(statuses, k)
			}
			slices.Sort(statuses)

			titles := make([]string, len(statusCols)+1)
			index := make([]int, len(statusCols)) // column substitution
			titles[0] = "Period start (UTC)"
			for i, status := range statuses {
				titles[i+1] = strconv.Itoa(status)
				index[statusCols[status]] = i
			}

			// convert matrix to output table, observing correct column index
			table := make([][]string, len(rsp.JSON200.Stats)+1)
			table[0] = titles
			for i := range rsp.JSON200.Stats {
				line := make([]string, len(statusCols)+1)
				line[0] = rsp.JSON200.Stats[i].Time.Format("2006-01-02T15:04:05")
				for j, count := range counts[i] {
					line[index[j]+1] = strconv.Itoa(count)
				}
				for j := len(counts[i]); j < len(index); j++ {
					line[index[j]+1] = "0"
				}
				table[i+1] = line
			}

			output.Table(table, output.Format(cmd))
			return nil
		},
	}
	statFlags(cmdCalls)

	var cmdDuration = &cobra.Command{
		Use:     "duration <app_name>",
		Aliases: []string{"duration", "time", "timing"},
		Short:   "Show app execution duration",
		Long: `Show duration of app calls, grouped by time slots. All times are in msec
By default it reports every hour from the beginning of current day (UTC),
but you can change reporting interval using "--from" and "--to" flags
(specifying date/time in format "YYYY-MM-DD HH:mm:SS", where either date or time,
can be omitted, or as UNIX timestamp) and reporting step duration with flag
"--step" (in seconds).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := getAppIdByName(args[0])
			if err != nil {
				return fmt.Errorf("cannot find app by name: %w", err)
			}

			from, err := parseTimeFlag(cmd, "from")
			if err != nil {
				return fmt.Errorf("cannot parse 'from' time: %w", err)
			}

			to, err := parseTimeFlag(cmd, "to")
			if err != nil {
				return fmt.Errorf("cannot parse 'to' time: %w", err)
			}

			step, err := cmd.Flags().GetInt("step")
			if err != nil {
				return fmt.Errorf("cannot parse reporting step: %w", err)
			}

			rsp, err := client.AppDurationWithResponse(
				context.Background(),
				id,
				&sdk.AppDurationParams{
					From: from,
					To:   to,
					Step: step,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot get statistics: %w", err)
			}
			if rsp.StatusCode() != http.StatusOK {
				return fmt.Errorf("cannot get statistics: %s", string(rsp.Body))
			}

			if output.Format(cmd) == output.FmtJSON {
				fmt.Println(string(rsp.Body))
				return nil
			}

			if len(rsp.JSON200.Stats) == 0 {
				fmt.Println("No data to report")
				return nil
			}

			table := make([][]string, len(rsp.JSON200.Stats)+1)
			table[0] = []string{"Period start (UTC)", "Min", "Avg", "Median", "75%", "90%", "Max"}
			for i, d := range rsp.JSON200.Stats {
				table[i+1] = []string{
					d.Time.Format("2006-01-02T15:04:05"),
					scaleToMsec(d.Min),
					scaleToMsec(d.Avg),
					scaleToMsec(d.Median),
					scaleToMsec(d.Perc75),
					scaleToMsec(d.Perc90),
					scaleToMsec(d.Max),
				}
			}

			output.Table(table, output.Format(cmd))
			return nil
		},
	}
	statFlags(cmdDuration)

	cmdStat.AddCommand(cmdCalls, cmdDuration)

	return cmdStat
}

func scaleToMsec(src int64) string {
	return fmt.Sprintf("%.0f", float64(src)/1000.0)
}

func statFlags(cmd *cobra.Command) {
	cmd.Flags().String("from", "today", "Reporting period start, UTC")
	cmd.Flags().String("to", "now", "Reporting period end, UTC")
	cmd.Flags().Int("step", 3600, "Reporting step, seconds")
}

func parseTimeFlag(cmd *cobra.Command, name string) (time.Time, error) {
	val, err := cmd.Flags().GetString(name)
	if err != nil {
		return time.Time{}, err
	}
	if val == "now" {
		return time.Now().UTC(), nil
	}
	if val == "today" {
		return carbon.Now().StartOfDay().StdTime(), nil
	}
	carb := carbon.Parse(val, carbon.UTC)
	if !carb.IsValid() {
		return time.Time{}, fmt.Errorf("cannot parse '%s' time", name)
	}
	return carb.StdTime(), nil
}
