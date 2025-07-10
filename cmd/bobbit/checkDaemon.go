package main

import (
	"mplus.software/oss/bobbit.go/client"
	"mplus.software/oss/bobbit.go/internal/shell"
)

func CheckDaemonHandler() {
	if _, err := client.CreateConnection(c); err != nil {
		shell.Fatalfln(3, "Error when checking daemon: %v", err)
	}
	shell.Println("Daemon is running.")
}
