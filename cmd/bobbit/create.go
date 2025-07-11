package main

import (
	"github.com/spf13/cobra"
	"mplus.software/oss/bobbit.go/client"
	"mplus.software/oss/bobbit.go/internal/shell"
	"mplus.software/oss/bobbit.go/payload"
)

func RegisterCreateCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "create <job_id> -- <command>",
		Short: "Create new job",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			id, command := args[0], args[1:]
			payload := payload.JobPayload{
				ID:      id,
				Command: command,
				Request: payload.EXECUTE_JOB,
			}

			conn, err := client.CreateConnection(c)
			if err != nil {
				shell.Fatalln(3, err.Error())
			}
			defer conn.Connection.Close()

			// TODO: Implement later
			//if _, err := os.Stat(c.GetLockfilePath(id)); err == nil {
			//	shell.Fatalfln(3, "Job %s is still running.", id)
			//}

			if err := conn.SendPayload(payload); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}
			shell.Printfln("Job %s created!", id)
		},
	})
}
