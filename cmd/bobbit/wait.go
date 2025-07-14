package main

import (
	"github.com/spf13/cobra"
	"mplus.software/oss/bobbit.go/client"
	"mplus.software/oss/bobbit.go/internal/shell"
	"mplus.software/oss/bobbit.go/payload"
)

func RegisterWaitCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "wait <jobID|jobName>",
		Short: "Wait for job.",
		Long:  "Wait for job. If user provide jobName that have same name, it will using the latest job.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := client.CreateConnection(c)
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}
			defer conn.Connection.Close()

			p := payload.JobPayload{Request: payload.REQUEST_WAIT}
			if err := p.MarshalMetadata(payload.JobSearchMetadata{Search: args[0]}); err != nil {
				shell.Fatalfln(3, "Failed to marshal metadata. %v", err)
			}

			if err := conn.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}

			var job payload.JobStatus
			if err := conn.GetPayload(&job); err != nil {
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
