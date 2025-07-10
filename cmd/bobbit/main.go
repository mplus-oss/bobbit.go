package main

import (
	"os"

	"github.com/spf13/cobra"
	"mplus.software/oss/bobbit.go/internal/config"
)

var (
	c   = config.New()
	cmd = &cobra.Command{
		Use:   "bobbit",
		Short: "Simply \"yet\" UNIX Socket based job runner",
	}
)

func init() {
	cmd.AddCommand(&cobra.Command{
		Use:   "create <job_id> -- <command>",
		Short: "Create new job",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			CreateCommandHandler(args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "is-running",
		Short: "Check if bobbit daemon is running.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			CheckDaemonHandler()
		},
	})
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
