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
			job, err := cli.Wait(args[0])
			if err != nil {
				shell.Fatalfln(3, "Failed to wait for job: %v", err)
			}

			status := payload.ParseJobStatus(job.Status)
			shell.Printfln("Job %s is finished with status \"%s\".", args[0], status)
		},
	})
}
