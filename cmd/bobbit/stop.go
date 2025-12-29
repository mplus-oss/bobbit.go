package main

import (
	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/spf13/cobra"
)

func RegisterStopCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "stop <job_name|id>",
		Short: "Stop running job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jobName := args[0]

			job, err := cli.Stop(jobName)
			if err != nil {
				shell.Fatalfln(3, "Failed to stop job: %v", err)
			}

			if job.ID == "" {
				shell.Fatalln(3, "No job found!")
			}
			shell.Printfln("Job %s [%s] has been stopped!", job.JobName, job.ID)
		},
	})
}
