package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/spf13/cobra"
)

func RegisterTailCommand() {
	tail := &cobra.Command{
		Use:   "tail <jobID|jobName>",
		Short: "Tail job log in real-time.",
		Long:  "Stream job log output in real-time. If user provide jobName that have same name, it will use the latest job.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			follow, err := cmd.Flags().GetBool("follow")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			err = cli.TailJobLogWithContext(ctx, args[0], follow, func(line string) error {
				fmt.Println(line)
				return nil
			})

			if err != nil && err != context.Canceled {
				shell.Fatalfln(3, "Failed to tail job log: %v", err)
			}
		},
	}

	tail.Flags().BoolP("follow", "f", false, "Follow log output (stream mode)")
	cmd.AddCommand(tail)
}
