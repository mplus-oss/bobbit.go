package main

import (
	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/mplus-oss/bobbit.go/payload"
	"github.com/spf13/cobra"
)

func RegisterWaitCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "wait <jobID|jobName>",
		Short: "Wait for job.",
		Long:  "Wait for job. If user provide jobName that have same name, it will using the latest job.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			p := payload.JobPayload{Request: payload.REQUEST_WAIT}
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

			if job.Status != payload.JOB_RUNNING {
				var status string
				switch job.Status {
				case payload.JOB_FAILED:
					status = "Failed"
				case payload.JOB_FINISH:
					status = "Finished"
				case payload.JOB_NOT_RUNNING:
					status = "Not running"
				}
				shell.Printfln("Job %s is finished with status \"%s\".", args[0], status)
			}
		},
	})
}
