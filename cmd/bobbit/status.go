package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/mplus-oss/bobbit.go/payload"
	"github.com/spf13/cobra"
)

func RegisterStatusCommand() {
	status := &cobra.Command{
		Use:   "status <jobID|jobName>",
		Short: "Check status from specific job.",
		Long:  "Check status from specific job. If user provide jobName that have same name, it will using the latest job.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showMetadata, err := cmd.Flags().GetBool("show-metadata")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}

			p := payload.JobPayload{Request: payload.REQUEST_STATUS}
			if err := cli.BuildPayload(&p, payload.JobSearchMetadata{Search: args[0]}); err != nil {
				shell.Fatalfln(3, "Failed to build payload: %v", err)
			}
			defer cli.Connection.Close()

			if err := cli.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}

			var job payload.JobResponse
			if err := cli.GetPayload(&job); err != nil {
				shell.Fatalfln(3, "Failed to get payload from daemon: %v", err)
			}

			var duration string
			timeStr := "%s [%s - %s]"
			now := time.Now()
			if job.Status == payload.JOB_RUNNING {
				duration = lib.HumanizeDuration(now.Sub(job.CreatedAt))
				timeStr = fmt.Sprintf(
					timeStr, "elapsed "+duration,
					job.CreatedAt.Local().String(),
					now.Local().String(),
				)
			} else {
				duration = lib.HumanizeDuration(job.UpdatedAt.Sub(job.CreatedAt))
				timeStr = fmt.Sprintf(
					timeStr, duration,
					job.CreatedAt.Local().String(),
					job.UpdatedAt.Local().String(),
				)
			}

			shell.Printf("Status for Job: %s\n", job.JobName)
			shell.Printf("------------------------\n")
			shell.Printf("  ID:        %s\n", job.ID)
			shell.Printf("  Status:    %s\n", payload.ParseJobStatus(job.Status))
			shell.Printf("  Exit Code: %d\n", job.ExitCode)
			shell.Printf("  Time:      %s\n", timeStr)

			if showMetadata && job.Metadata != nil {
				metaBytes, err := json.MarshalIndent(job.Metadata, "", "  ")
				if err == nil {
					shell.Printf("  Metadata:\n%s\n", string(metaBytes))
				}
			}
		},
	}
	status.Flags().Bool("show-metadata", false, "Show metadata")
	cmd.AddCommand(status)
}
