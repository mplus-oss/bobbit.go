package main

import (
	"mplus.software/oss/bobbit.go/client"
	"mplus.software/oss/bobbit.go/internal/shell"
	"mplus.software/oss/bobbit.go/payload"
)

func CreateCommandHandler(args []string) {
	id, command := args[0], args[1:]

	conn, err := client.CreateConnection(c)
	if err != nil {
		shell.Fatalln(3, err.Error())
	}
	defer conn.Connection.Close()

	// TODO: Implement later
	//if _, err := os.Stat(c.GetLockfilePath(id)); err == nil {
	//	shell.Fatalfln(3, "Job %s is still running.", id)
	//}

	if err := conn.SendPayload(payload.JobPayload{ID: id, Command: command}); err != nil {
		shell.Fatalfln(3, "Failed to send payload to daemon: %v", err)
	}
	shell.Printfln("Job %s created!", id)
}
