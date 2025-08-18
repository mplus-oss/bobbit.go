package main

import (
	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/mplus-oss/bobbit.go/payload"
	"github.com/spf13/cobra"
)

func RegisterStopCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "stop <job_name|id>",
		Short: "Stop running job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jobName := args[0]

			p := payload.JobPayload{Request: payload.REQUEST_STOP}
			req := payload.JobSearchMetadata{
				Search: jobName,
			}
			if err := cli.BuildPayload(&p, req); err != nil {
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

			if job.ID == "" {
				shell.Fatalln(3, "No job found!")
			}
			shell.Printfln("Job %s [%s] has been stopped!", job.JobName, job.ID)
		},
	})
}
