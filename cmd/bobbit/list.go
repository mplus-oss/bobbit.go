package main

import (
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

			p := payload.JobPayload{Request: payload.REQUEST_LIST}
			req := payload.JobSearchMetadata{
				RequestMeta: false,
				ActiveOnly:  activeOnly,
			}
			if err := cli.BuildPayload(&p, req); err != nil {
				shell.Fatalfln(3, "Failed to build payload: %v", err)
			}

			if err := cli.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}

			var jobs []payload.JobResponse
			if err := cli.GetPayload(&jobs); err != nil {
				shell.Fatalfln(3, "Failed to get payload from daemon: %v", err)
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
	list.Flags().Bool("active-only", false, "List only for job with running/active status")
	cmd.AddCommand(list)
}
