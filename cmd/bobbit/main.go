package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"mplus.software/oss/bobbit.go/internal/config"
	"mplus.software/oss/bobbit.go/internal/daemon"
	"mplus.software/oss/bobbit.go/internal/shell"
)

var (
	c   = config.New()
	cmd = &cobra.Command{
		Use:   "bobbit",
		Short: "Simply \"yet\" FIFO based job runner",
	}
)

func init() {
	cmd.AddCommand(&cobra.Command{
		Use:   "create <job_id> <command>",
		Short: "Create new job",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			id, command := args[0], args[1:]

			conn, err := daemon.CreateConnection(c)
			if err != nil {
				shell.Fatalln(3, err.Error())
			}
			defer conn.Connection.Close()

			lockFile := filepath.Join(c.DataDir, id+".lock")
			if _, err := os.Stat(lockFile); err == nil {
				shell.Fatalfln(3, "Job %s is still running.", id)
			}

			if err := conn.SendPayload(config.JobPayload{ID: id, Command: command}); err != nil {
				shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
			}
			shell.Printfln("Job %s created!", id)
		},
	})
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
