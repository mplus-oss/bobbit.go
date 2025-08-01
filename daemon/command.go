package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/payload"
)

// HandleVibeCheck handles a "vibe check" request, which typically serves as a basic
// ping to confirm the daemon is responsive. It unmarshals the request metadata.
func (d *DaemonStruct) HandleVibeCheck(jc *JobContext) error {
	var payload payload.PayloadRegularMetadata
	if err := jc.Payload.UnmarshalMetadata(&payload); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}
	return nil
}

// HandleJob processes a new job request. It validates the job payload,
// generates a unique ID if not provided, creates necessary data files (lock, log, metadata),
// executes the command, captures its output and exit code, and cleans up the lock file.
func (d *DaemonStruct) HandleJob(jc *JobContext) error {
	var payload payload.JobDetailMetadata
	if err := jc.Payload.UnmarshalMetadata(&payload); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	if payload.JobName == "" || len(payload.Command) < 1 {
		return &DaemonError{"Invalid payload: JobName or Command not provided", nil}
	}
	if payload.ID == "" {
		hash, err := lib.GenerateRandomHash(16)
		if err != nil {
			return &DaemonError{"Failed to create Hash for job", err}
		}
		payload.ID = hash
	}
	if payload.Timestamp.IsZero() {
		payload.Timestamp = jc.Payload.Timestamp
	}

	lockFile := GenerateJobDataFilename(d.BobbitConfig, payload, DAEMON_LOCKFILE)
	logFile := GenerateJobDataFilename(d.BobbitConfig, payload, DAEMON_LOGFILE)
	exitCodeFile := GenerateJobDataFilename(d.BobbitConfig, payload, DAEMON_EXITCODE)
	metadataFile := GenerateJobDataFilename(d.BobbitConfig, payload, DAEMON_METADATA)

	log.Printf("Entering HandleJob Context: %v", jc)
	if err := os.WriteFile(lockFile, []byte{}, 0644); err != nil {
		return &DaemonPayloadError{"Failed to create lockfile", payload.ID, err}
	}
	defer os.Remove(lockFile)

	metadataStr := ""
	if payload.Metadata != nil {
		metaByte, err := json.Marshal(payload.Metadata)
		if err != nil {
			return &DaemonPayloadError{"Failed to marshal metadata", payload.ID, err}
		}

		if err := os.WriteFile(metadataFile, metaByte, 0644); err != nil {
			return &DaemonPayloadError{"Failed to create metadata file", payload.ID, err}
		}

		metadataStr = string(metaByte)
	}

	logOutput, err := os.Create(logFile)
	if err != nil {
		return &DaemonPayloadError{"Failed to create logfile", payload.ID, err}
	}
	defer logOutput.Close()

	exitCode := 0
	if len(payload.Command) == 0 {
		return &DaemonPayloadError{"No command provided", payload.ID, err}
	}

	cmd := exec.Command(payload.Command[0], payload.Command[1:]...)
	cmd.Stdout = logOutput
	cmd.Stderr = logOutput

	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("JOB_ID=%s", payload.ID),
		fmt.Sprintf("JOB_NAME=%s", payload.JobName),
		fmt.Sprintf("JOB_METADATA=%s", metadataStr),
	)

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
	log.Printf("DONE: %s %s exit=%d", payload.ID, payload.JobName, exitCode)
	return nil
}

// ListJob handles requests to list jobs. It reads job data from the configured
// directory, filters them based on `JobSearchMetadata` criteria (e.g., active only, limit),
// parses their status and optional metadata, sorts them, and sends the results back to the client.
//
// Check payload.JobSearchMetadata for more information.
func (d *DaemonStruct) ListJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobSearchMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	files, err := os.ReadDir(d.DataDir)
	if err != nil {
		return &DaemonError{"Failed to read bobbit directory", err}
	}

	log.Printf("Entering ListJob Context: %v", jc)
	jobIDs := make(map[string]bool)
	for _, file := range files {
		if filestat, err := os.Stat(filepath.Join(d.DataDir, file.Name())); err != nil {
			log.Printf("Error when checking status of file: %v", err)
			continue
		} else {
			if filestat.IsDir() {
				continue
			}
		}

		jobfile := SplitFilenameFromExtfile(file.Name())
		jobIDs[jobfile] = true
	}

	allJobs := []payload.JobResponse{}
	for id := range jobIDs {
		if statusRequest.FinishOnly && statusRequest.ActiveOnly {
			continue
		}

		metadata, err := ParseJobDataFilename(id)
		if err != nil {
			log.Println(err)
			continue
		}
		status := payload.JobResponse{JobDetailMetadata: metadata}
		if err := ParseExitCode(d.BobbitConfig, &status); err != nil {
			log.Printf("Failed to parse exit code from %s job: %v\n", id, err)
			continue
		}

		if statusRequest.RequestMeta {
			if metaBytes, err := os.ReadFile(GenerateJobDataFilename(d.BobbitConfig, metadata, DAEMON_METADATA)); err == nil {
				err := json.Unmarshal(metaBytes, &status.Metadata)
				if err != nil {
					log.Printf("Failed to unmarshal metadata from %s job: %v\n", id, err)
					continue
				}
			}
		}

		if statusRequest.FinishOnly && status.Status == payload.JOB_RUNNING {
			continue
		}

		if statusRequest.ActiveOnly && status.Status != payload.JOB_RUNNING {
			continue
		}

		allJobs = append(allJobs, status)
		if statusRequest.Limit > 0 && len(allJobs) >= statusRequest.Limit {
			break
		}
	}
	sort.Slice(allJobs, func(i, j int) bool {
		if statusRequest.OrderDesc {
			return allJobs[i].Timestamp.Before(allJobs[j].Timestamp)
		} else {
			return allJobs[j].Timestamp.Before(allJobs[i].Timestamp)
		}
	})

	if statusRequest.NumberOnly {
		if err := jc.SendPayload(payload.JobResponseCount{Count: len(allJobs)}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	} else {
		if err := jc.SendPayload(allJobs); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	}

	log.Println("DONE: LIST")
	return nil
}

// WaitJob handles requests to wait for a specific job to complete.
// It continuously checks for the absence of the job's lock file and
// responds to the client with the job's final status once completed.
// It also handles client connection timeouts.
func (d *DaemonStruct) WaitJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobSearchMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	job, err := FindJobDataFilename(d.BobbitConfig, statusRequest)
	if err != nil {
		return &DaemonError{"Failed to find job data", err}
	}

	if job.ID == "" {
		if err := jc.SendPayload(payload.JobResponse{Status: payload.JOB_NOT_RUNNING}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	}

	lockFile := GenerateJobDataFilename(d.BobbitConfig, job, DAEMON_LOCKFILE)
	log.Printf("Entering WaitJob Context: %v\n", job)
	for {
		oneByte := make([]byte, 1)
		jc.conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
		if _, err := jc.conn.Read(oneByte); err != io.EOF {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				log.Printf("Connection closed from client. job=%v\n", job)
				return err
			}
		} else {
			return &DaemonError{"Connection Error", err}
		}
		jc.conn.SetReadDeadline(time.Time{})

		if _, err := os.Stat(lockFile); os.IsNotExist(err) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	resp := payload.JobResponse{JobDetailMetadata: job}
	if err := ParseExitCode(d.BobbitConfig, &resp); err != nil {
		return &DaemonPayloadError{"Failed to parse exit code", job.JobName, err}
	}
	if err := jc.SendPayload(resp); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

	log.Printf("Job waiting finish: %v\n", job)
	return nil
}

// StatusJob handles requests to retrieve the current status of a specific job.
// It finds the job based on the provided search metadata, parses its exit code
// to determine its status, and sends the `JobResponse` back to the client.
func (d *DaemonStruct) StatusJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobSearchMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	job, err := FindJobDataFilename(d.BobbitConfig, statusRequest)
	if err != nil {
		return &DaemonError{"Failed to find job data", err}
	}

	if job.ID == "" {
		if err := jc.SendPayload(payload.JobResponse{Status: payload.JOB_NOT_RUNNING}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	}

	finalStatus := payload.JobResponse{JobDetailMetadata: job}
	if err := ParseExitCode(d.BobbitConfig, &finalStatus); err != nil {
		return &DaemonPayloadError{"Failed to parse exit code", job.JobName, err}
	}
	if err := jc.SendPayload(finalStatus); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

	return nil
}
