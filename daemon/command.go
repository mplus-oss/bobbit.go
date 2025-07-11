package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"mplus.software/oss/bobbit.go/payload"
)

func (d *DaemonStruct) HandleJob(jc *JobContext) error {
	var payload payload.JobRequestMetadata
	if err := jc.Payload.UnmarshalMetadata(&payload); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata: %v", err}
	}

	if payload.ID == "" || len(payload.Command) < 1 {
		return &DaemonError{"Invalid payload: ID or Command not provided.", nil}
	}

	jobPath := func(ext string) string {
		return filepath.Join(
			d.DataDir,
			fmt.Sprintf("%s-%s.%s",
				jc.Payload.Timestamp.Format(time.RFC3339),
				payload.ID,
				ext,
			),
		)
	}
	lockFile := jobPath("lock")
	logFile := jobPath("log")
	exitCodeFile := jobPath("exitcode")
	metadataFile := jobPath("metadata")

	if err := os.WriteFile(lockFile, []byte{}, 0644); err != nil {
		return &DaemonPayloadError{"Failed to create lockfile.", payload.ID, err}
	}
	defer os.Remove(lockFile)

	if payload.Metadata != nil {
		metaByte, err := json.Marshal(payload.Metadata)
		if err != nil {
			return &DaemonPayloadError{"Failed to marshal metadata.", payload.ID, err}
		}
		if err := os.WriteFile(metadataFile, metaByte, 0644); err != nil {
			return &DaemonPayloadError{"Failed to create metadata file.", payload.ID, err}
		}
	}

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

func (d *DaemonStruct) ListJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobStatusMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return err
	}

	files, err := os.ReadDir(d.DataDir)
	if err != nil {
		return err
	}

	jobIDs := make(map[string]bool)
	for _, file := range files {
		parts := strings.Split(file.Name(), ".")
		if len(parts) > 1 {
			jobIDs[parts[0]] = true
		}
	}

	allJobs := []payload.JobStatus{}
	for id := range jobIDs {
		jobPath := func(ext string) string { return filepath.Join(d.DataDir, id+ext) }

		status := payload.JobStatus{}
		status.ID = strings.Join(strings.Split(id, "-")[3:], "-")

		timestamp, err := time.Parse(time.RFC3339, strings.Join(strings.Split(id, "-")[:3], "-"))
		if err != nil {
			log.Println(err)
			continue
		}
		status.Timestamp = timestamp

		if statusRequest.RequestMeta {
			if metaBytes, err := os.ReadFile(jobPath(".metadata")); err == nil {
				json.Unmarshal(metaBytes, &status.Metadata)
			}
		}
		if _, err := os.Stat(jobPath(".lock")); err == nil {
			status.Status = payload.JOB_RUNNING
		}
		if exitCodeBytes, err := os.ReadFile(jobPath(".exitcode")); err == nil {
			code, _ := strconv.Atoi(strings.TrimSpace(string(exitCodeBytes)))
			status.ExitCode = code
			if code == 0 {
				status.Status = payload.JOB_FINISH
			} else {
				status.Status = payload.JOB_FAILED
			}
		}

		allJobs = append(allJobs, status)
	}
	sort.Slice(allJobs, func(i, j int) bool {
		return allJobs[i].Timestamp.Before(allJobs[j].Timestamp)
	})

	if err := jc.SendPayload(allJobs); err != nil {
		return err
	}

	log.Println("DONE: LIST")
	return nil
}
