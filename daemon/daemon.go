package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"mplus.software/oss/bobbit.go/config"
	"mplus.software/oss/bobbit.go/payload"
)

type DaemonStruct struct {
	SocketListener net.Listener
	config.BobbitDaemonConfig
}

func CreateDaemon(c config.BobbitDaemonConfig) (*DaemonStruct, error) {
	if err := os.MkdirAll(c.DataDir, 0755); err != nil {
		return nil, &DaemonError{"Failed to create data directory: %v", err}
	}

	if err := os.RemoveAll(c.SocketPath); err != nil {
		return nil, &DaemonError{"Failed to remove old socket path: %v", err}
	}

	listener, err := net.Listen("unix", c.SocketPath)
	if err != nil {
		return nil, &DaemonError{"Failed to listen in socket path: %v", err}
	}

	return &DaemonStruct{
		SocketListener:     listener,
		BobbitDaemonConfig: c,
	}, nil
}

func (d *DaemonStruct) CleanupDaemon(sigChan <-chan os.Signal) {
	<-sigChan
	log.Println("Cleanup daemon...")
	os.Remove(d.SocketPath)
	os.Exit(0)
}

func (d *DaemonStruct) GetPayload(conn net.Conn) (payload payload.JobPayload, err error) {
	if err = json.NewDecoder(conn).Decode(&payload); err != nil {
		return payload, &DaemonError{"Failed to decode payload.", err}
	}
	if payload.ID == "" || len(payload.Command) < 1 {
		return payload, &DaemonError{"Invalid payload: ID or Command not provided.", err}
	}

	return payload, nil
}

func (d *DaemonStruct) HandleJob(payload payload.JobPayload) error {
	lockFile := d.BobbitDaemonConfig.GetLockfilePath(payload.ID)
	logFile := d.BobbitDaemonConfig.GetLogfilePath(payload.ID)
	exitCodeFile := d.BobbitDaemonConfig.GetExitCodePath(payload.ID)

	if err := os.WriteFile(lockFile, []byte{}, 0644); err != nil {
		return &DaemonPayloadError{"Failed to create lockfile.", payload.ID, err}
	}
	defer os.Remove(lockFile)

	logOutput, err := os.Create(logFile)
	if err != nil {
		return &DaemonPayloadError{"Failed to create logfile.", payload.ID, err}
	}
	defer logOutput.Close()

	exitCode := 0
	if len(payload.Command) == 0 {
		return &DaemonPayloadError{"No command provided.", payload.ID, err}
	}

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

	if err := os.WriteFile(exitCodeFile, fmt.Appendf([]byte{}, "%d", exitCode), 0644); err != nil {
		(&DaemonPayloadError{"Failed to create exitcode file. Please handle it manually.", payload.ID, err}).Warning()
	}
	log.Printf("DONE: %s exit=%d", payload.ID, exitCode)
	return nil
}
