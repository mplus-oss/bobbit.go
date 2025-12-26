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
			finishOnly, err := cmd.Flags().GetBool("finish-only")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			orderDesc, err := cmd.Flags().GetBool("desc")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			countMaxJob, err := cmd.Flags().GetInt("count")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			pageJob, err := cmd.Flags().GetInt("page")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			jobNumberOnly, err := cmd.Flags().GetBool("total")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			toJson, err := cmd.Flags().GetBool("to-json")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			metadataFilter, err := cmd.Flags().GetStringToString("metadata")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}

			p := payload.JobPayload{Request: payload.REQUEST_LIST}
			req := payload.JobSearchMetadata{
				RequestMeta:    false,
				ActiveOnly:     activeOnly,
				Limit:          countMaxJob,
				Page:           pageJob,
				NumberOnly:     jobNumberOnly,
				OrderDesc:      orderDesc,
				FinishOnly:     finishOnly,
				MetadataFilter: metadataFilter,
			}
			if err := cli.BuildPayload(&p, req); err != nil {
				shell.Fatalfln(3, "Failed to build payload: %v", err)
			}

			if err := cli.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}

			if jobNumberOnly {
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
			fmt.Fprintln(w, "Start\tUpdate\tID (Short)\tName\tStatus\tExit Code")
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
					w, "%v\t%v\t%s\t%s\t%s\t%d\n",
					job.CreatedAt.Format(time.RFC3339),
					job.UpdatedAt.Format(time.RFC3339),
					job.ID[:16],
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

	list.Flags().BoolP("active-only", "a", false, "Filters the list to show only jobs with a running or active status")
	list.Flags().BoolP("finish-only", "f", false, "Filters the list to show only jobs with a finish or failed status")
	list.Flags().Bool("desc", false, "Orders the list of jobs in descending order")
	list.Flags().IntP("count", "n", 0, "Sets a maximum number of jobs to return")
	list.Flags().IntP("page", "p", 0, "Create pagination of jobs based on limit option")
	list.Flags().Bool("total", false, "Returns only the total count of jobs instead of the full list")
	list.Flags().BoolP("to-json", "j", false, "Print the list to stringify JSON")
	list.Flags().StringToStringP("metadata", "m", nil, "Filter jobs by metadata (e.g., -m 'key1=value1,key2=%%value2%%')")

	cmd.AddCommand(list)
}
