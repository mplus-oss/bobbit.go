package main

import (
	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/mplus-oss/bobbit.go/payload"
	"github.com/spf13/cobra"
)

func RegisterDaemonCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "is-running",
		Short: "Check if bobbit daemon is running.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			p := payload.JobPayload{Request: payload.REQUEST_VIBE_CHECK}
			if err := cli.BuildPayload(&p, make(map[string]string, 1)); err != nil {
				shell.Fatalfln(3, "Failed to build payload: %v", err)
			}
			defer cli.Connection.Close()

			if err := cli.SendPayload(p); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}
			shell.Println("Daemon is running.")
		},
	})
}
