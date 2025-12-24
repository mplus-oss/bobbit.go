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
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/metadata/models"
	"github.com/mplus-oss/bobbit.go/payload"
)

type HandlerFunc func(jc *JobContext) error

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
	var p payload.JobDetailMetadata
	if err := jc.Payload.UnmarshalMetadata(&p); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	if p.JobName == "" || len(p.Command) < 1 {
		return &DaemonError{"Invalid p: JobName or Command not provided", nil}
	}
	// Generate a unique ID if not provided
	if p.ID == "" {
		hash, err := lib.GenerateRandomHash(32)
		if err != nil {
			return &DaemonError{"Failed to create Hash for job", err}
		}
		p.ID = hash
	}

	// Marshal and write job metadata
	metadataStr := ""
	if p.Metadata != nil {
		metaByte, err := json.Marshal(p.Metadata)
		if err != nil {
			return &DaemonPayloadError{"Failed to marshal metadata", p.ID, err}
		}

		metadataStr = string(metaByte)
	}

	// Prepare the logfile
	logFile := GenerateJobLogPath(jc.daemon.BobbitDaemonConfig, p)
	logOutput, err := os.Create(logFile)
	if err != nil {
		return &DaemonPayloadError{"Failed to create logfile", p.ID, err}
	}
	defer logOutput.Close()

	if len(p.Command) == 0 {
		return &DaemonPayloadError{"No command provided", p.ID, err}
	}

	// Save the process
	respPayload := payload.JobResponse{
		ExitCode:          -1,
		Status:            payload.JOB_NOT_RUNNING,
		JobDetailMetadata: p,
	}
	job, err := models.NewJobModel(jc.daemon.DB, respPayload)
	if err != nil {
		return &DaemonPayloadError{"Failed when initialize db model", p.ID, err}
	}
	if err := job.Save(); err != nil {
		return &DaemonPayloadError{"Failed when creating job record", p.ID, err}
	}

	// Prep the output
	cmd := exec.Command(p.Command[0], p.Command[1:]...)
	cmd.Stdout = logOutput
	cmd.Stderr = logOutput

	// Make it as a different group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Add job-specific environment variables
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("JOB_ID=%s", p.ID),
		fmt.Sprintf("JOB_NAME=%s", p.JobName),
		fmt.Sprintf("JOB_METADATA=%s", metadataStr),
	)

	log.Printf("Starting Job: %+v", p)
	if err := cmd.Start(); err != nil {
		job.Delete()
		return &DaemonPayloadError{"Failed when starting command", p.ID, err}
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			job.ExitCode = exitErr.ExitCode()
		} else {
			job.ExitCode = 127
		}
	} else {
		job.ExitCode = 0
	}

	if err := job.MarkJobFinished(); err != nil {
		return &DaemonPayloadError{"Failed to update job", p.ID, err}
	}

	if job.ExitCode > 0 {
		return &DaemonPayloadError{"Exit with code", p.ID, fmt.Errorf("code: %d", job.ExitCode)}
	}

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

	jobIDs := make(map[string]bool)
	// Collect unique job IDs from the filenames
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

	// Convert map keys to slice and sort
	jobIDSlice := make([]string, 0, len(jobIDs))
	for k := range jobIDs {
		jobIDSlice = append(jobIDSlice, k)
		delete(jobIDs, k)
	}
	slices.Sort(jobIDSlice)
	// Reverse order if requested
	if statusRequest.OrderDesc {
		slices.Reverse(jobIDSlice)
	}

	// Apply pagination (limit and page) if specified
	if statusRequest.Limit > 0 {
		start := 0
		end := statusRequest.Limit

		if statusRequest.Page > 0 {
			start = (statusRequest.Page - 1) * statusRequest.Limit
			end = start + statusRequest.Limit
		}

		if start >= len(jobIDSlice) {
			jobIDSlice = []string{}
		} else {
			end = min(end, len(jobIDSlice))
			jobIDSlice = jobIDSlice[start:end]
		}
	}

	allJobs := []payload.JobResponse{}
	// Iterate through selected job IDs to gather full job details
	for _, id := range jobIDSlice {
		// Skip if both finish-only and active-only filters are set (mutually exclusive)
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
	}

	// Send back job count or full job details based on request
	if statusRequest.NumberOnly {
		jobs := len(allJobs)
		allJobs = nil
		if err := jc.SendPayload(payload.JobResponseCount{Count: jobs}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	} else {
		if err := jc.SendPayload(allJobs); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	}

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

	// If job ID is empty, the job was not found, send a "not running" status
	if job.ID == "" {
		if err := jc.SendPayload(payload.JobResponse{Status: payload.JOB_NOT_RUNNING}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
	}

	lockFile := GenerateJobDataFilename(d.BobbitConfig, job, DAEMON_LOCKFILE)
	for {
		// Set a short read deadline to check for client disconnection without blocking indefinitely
		oneByte := make([]byte, 1)
		jc.conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
		if _, err := jc.conn.Read(oneByte); err != io.EOF {
			// If error is not EOF and not a timeout, it means connection issue, return error
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				log.Printf("Connection closed from client. job=%v\n", job)
				return err
			}
		} else {
			// If EOF is returned, the client has closed the connection
			return &DaemonError{"Connection Error", err}
		}
		jc.conn.SetReadDeadline(time.Time{})

		// Check if the lock file exists; if not, the job has completed
		if _, err := os.Stat(lockFile); os.IsNotExist(err) {
			break
		}
		// Wait before polling again to avoid busy-waiting
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

	// Send FIN/Half-Close Connection
	if unixConn, ok := jc.conn.(*net.UnixConn); ok {
		if err := unixConn.CloseWrite(); err != nil {
			log.Printf("Warning: Failed to CloseWrite the job: %v\n", err)
		}
	}
	if _, err = io.Copy(io.Discard, jc.conn); err != nil {
		log.Printf("EOF Connection: %v\n", err)
	}
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

// StopJob handles requests to stop the specific job.
// If the job exist, the return is JobResponse. If the job not exist, the return is an empty JobResponse.
func (d *DaemonStruct) StopJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobSearchMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	job, err := FindJobDataFilename(d.BobbitConfig, statusRequest)
	if err != nil {
		return &DaemonError{"Failed to find job data", err}
	}

	resp := payload.JobResponse{JobDetailMetadata: job}
	if err := ParseExitCode(d.BobbitConfig, &resp); err != nil {
		return &DaemonError{"Failed to parse exitcode", err}
	}

	// Send empty response
	if resp.Status != payload.JOB_RUNNING {
		if err := jc.SendPayload(payload.JobResponse{}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}
		return nil
	}

	pidBytes, err := os.ReadFile(GenerateJobDataFilename(d.BobbitConfig, job, DAEMON_LOCKFILE))
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return err
	}

	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		return err
	}

	if err := jc.SendPayload(resp); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

	return nil
}
