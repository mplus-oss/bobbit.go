package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/mplus-oss/bobbit.go/payload"
	"github.com/spf13/cobra"
)

func RegisterListCommand() {
	list := &cobra.Command{
		Use:   "list",
		Short: "List of job",
		Run: func(cmd *cobra.Command, args []string) {
			activeOnly, err := cmd.Flags().GetBool("active-only")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			orderDesc, err := cmd.Flags().GetBool("order-desc")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			limitJob, err := cmd.Flags().GetInt("limit")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			countJob, err := cmd.Flags().GetBool("count")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			toJson, err := cmd.Flags().GetBool("to-json")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}

			p := payload.JobPayload{Request: payload.REQUEST_LIST}
			req := payload.JobSearchMetadata{
				RequestMeta: false,
				ActiveOnly:  activeOnly,
				Limit:       limitJob,
				NumberOnly:  countJob,
				OrderDesc:   orderDesc,
			}
			if err := cli.BuildPayload(&p, req); err != nil {
				shell.Fatalfln(3, "Failed to build payload: %v", err)
			}

			if err := cli.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}

			if countJob {
				var resp payload.JobResponseCount
				if err := cli.GetPayload(&resp); err != nil {
					shell.Fatalfln(3, "Failed to get payload from daemon: %v", err)
				}
				shell.Printfln("%v", resp.Count)
				return
			}

			var jobs []payload.JobResponse
			if err := cli.GetPayload(&jobs); err != nil {
				shell.Fatalfln(3, "Failed to get payload from daemon: %v", err)
			}

			if toJson {
				byteStr, err := json.Marshal(jobs)
				if err != nil {
					shell.Fatalln(3, err.Error())
					return
				}
				shell.Println(string(byteStr))
				return
			}

			w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "Time\tID\tName\tStatus\tExit Code")
			for _, job := range jobs {
				var status string
				switch job.Status {
				case payload.JOB_FAILED:
					status = "Failed"
				case payload.JOB_FINISH:
					status = "Finished"
				case payload.JOB_RUNNING:
					status = "Running"
				}

				fmt.Fprintf(
					w, "%v\t%s\t%s\t%s\t%d\n",
					job.Timestamp.Format(time.RFC3339),
					job.ID,
					job.JobName,
					status,
					job.ExitCode,
				)
			}
			if err := w.Flush(); err != nil {
				shell.Fatalfln(3, "Failed to print table: %v", err)
			}
		},
	}

	list.Flags().Bool("active-only", false, "Filters the list to show only jobs with a running or active status")
	list.Flags().Bool("order-desc", false, "Orders the list of jobs in descending order")
	list.Flags().Int("limit", 0, "Sets a maximum number of jobs to return")
	list.Flags().Bool("count", false, "Returns only the total count of jobs instead of the full list")
	list.Flags().BoolP("to-json", "j", false, "Print the list to stringify JSON")

	cmd.AddCommand(list)
}
