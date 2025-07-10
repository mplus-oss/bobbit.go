package main

import (
	"os"

	"mplus.software/oss/bobbit.go/internal/config"
	"mplus.software/oss/bobbit.go/internal/daemon"
	"mplus.software/oss/bobbit.go/internal/shell"
)

func CreateCommandHandler(args []string) {
	id, command := args[0], args[1:]

	conn, err := daemon.CreateConnection(c)
	if err != nil {
		shell.Fatalln(3, err.Error())
	}
	defer conn.Connection.Close()

	if _, err := os.Stat(c.GetLockfilePath(id)); err == nil {
		shell.Fatalfln(3, "Job %s is still running.", id)
	}

	if err := conn.SendPayload(config.JobPayload{ID: id, Command: command}); err != nil {
		shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
	}
	shell.Printfln("Job %s created!", id)
}
