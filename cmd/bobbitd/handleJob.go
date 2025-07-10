package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"mplus.software/oss/bobbit.go/internal/config"
)

func HandleJob(payload config.JobPayload) {
	lockFile := c.GetLockfilePath(payload.ID)
	logFile := c.GetLogfilePath(payload.ID)
	exitCodeFile := c.GetExitCodePath(payload.ID)

	if err := os.WriteFile(lockFile, []byte{}, 0644); err != nil {
		log.Printf("ERROR: Failed to create lockfile for job %s: %v", payload.ID, err)
		return
	}
	defer os.Remove(lockFile)

	logOutput, err := os.Create(logFile)
	if err != nil {
		log.Printf("ERROR: Failed to create logfile for job %s: %v", payload.ID, err)
		return
	}
	defer logOutput.Close()

	exitCode := 0
	if len(payload.Command) < 1 {
		log.Printf("ERROR: Command string for job %s is empty", payload.ID)
		exitCode = 127
	} else {
		cmd := exec.Command(payload.Command[0], payload.Command[1:]...)
		cmd.Stdout = logOutput
		cmd.Stderr = logOutput
		cmd.Env = append(os.Environ(), fmt.Sprintf("JOB_ID=%s", payload.ID))
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 127
			}
		}
	}

	if err := os.WriteFile(exitCodeFile, fmt.Appendf([]byte{}, "%d", exitCode), 0644); err != nil {
		log.Printf("ERROR: Failed to create exitcode file for job %s: %v", payload.ID, err)
	}
	log.Printf("DONE: %s exit=%d", payload.ID, exitCode)
}
