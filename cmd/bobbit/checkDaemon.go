package main

import (
	"mplus.software/oss/bobbit.go/internal/daemon"
	"mplus.software/oss/bobbit.go/internal/shell"
)

func CheckDaemonHandler() {
	if _, err := daemon.CreateConnection(c); err != nil {
		shell.Fatalfln(3, "Error when checking daemon: %v", err)
	}
	shell.Println("Daemon is running.")
}
