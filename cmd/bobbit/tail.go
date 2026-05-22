package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/internal/shell"
	"github.com/nxadm/tail"
	"github.com/spf13/cobra"
)

func RegisterTailCommand() {
	tailCmd := &cobra.Command{
		Use:   "tail <jobID|jobName>",
		Short: "Tail job log in real-time.",
		Long:  "Stream job log output in real-time. If user provide jobName that have same name, it will use the latest job.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			follow, err := cmd.Flags().GetBool("follow")
			if err != nil {
				shell.Fatalfln(3, "%v", err)
			}

			direct, err := cmd.Flags().GetBool("direct")
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

			if direct {
				err = tailDirectly(ctx, args[0], follow)
			} else {
				err = cli.TailJobLogWithContext(ctx, args[0], follow, func(line string) error {
					fmt.Println(line)
					return nil
				})
			}

			if err != nil && err != context.Canceled {
				shell.Fatalfln(3, "Failed to tail job log: %v", err)
			}
		},
	}

	tailCmd.Flags().BoolP("follow", "f", false, "Follow log output (stream mode)")
	tailCmd.Flags().BoolP("direct", "d", false, "Tail log file directly from filesystem instead of streaming over socket")
	cmd.AddCommand(tailCmd)
}

func tailDirectly(ctx context.Context, jobIDOrName string, follow bool) error {
	job, err := cli.Status(jobIDOrName)
	if err != nil {
		return fmt.Errorf("failed to get job status: %w", err)
	}

	c := config.BobbitConfig{
		DataPath:   cli.DataPath,
		SocketPath: cli.SocketPath,
		DebugMode:  cli.DebugMode,
	}
	logPath := config.GenerateJobLogPath(c, job.JobDetailMetadata)
	if logPath == "" {
		return fmt.Errorf("failed to resolve log path for job %s", jobIDOrName)
	}

	t, err := tail.TailFile(logPath, tail.Config{
		Follow:    follow,
		ReOpen:    follow,
		MustExist: true,
		Poll:      true,
	})
	if err != nil {
		return fmt.Errorf("failed to tail log file: %w", err)
	}
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case line, ok := <-t.Lines:
			if !ok {
				return nil
			}
			if line.Err != nil {
				return fmt.Errorf("error reading log line: %w", line.Err)
			}
			fmt.Println(line.Text)
		}
	}
}
