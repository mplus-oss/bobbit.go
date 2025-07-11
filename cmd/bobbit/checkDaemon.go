package main

import (
	"github.com/spf13/cobra"
	"mplus.software/oss/bobbit.go/client"
	"mplus.software/oss/bobbit.go/internal/shell"
)

func RegisterDaemonCommand() {
	cmd.AddCommand(&cobra.Command{
		Use:   "is-running",
		Short: "Check if bobbit daemon is running.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := client.CreateConnection(c); err != nil {
				shell.Fatalfln(3, "Error when checking daemon: %v", err)
			}
			shell.Println("Daemon is running.")
		},
	})
}
