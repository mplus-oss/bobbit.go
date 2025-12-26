package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"syscall"

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

	// Set timestamp if not provided. This is for logfile path.
	if p.CreatedAt.IsZero() {
		p.CreatedAt = jc.Payload.Timestamp
		p.UpdatedAt = jc.Payload.Timestamp
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

	job.Status = int(payload.JOB_RUNNING)
	job.PID = cmd.Process.Pid
	if err := job.Update(); err != nil {
		log.Printf("[WARNING] Failed when updating status: %+v", err)
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
	var req payload.JobSearchMetadata
	if err := jc.Payload.UnmarshalMetadata(&req); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	job, err := models.NewJobModel(jc.daemon.DB, payload.JobResponse{})
	if err != nil {
		return &DaemonError{"Failed when initialize db model", err}
	}

	filter := &models.JobFilter{
		ActiveOnly:     req.ActiveOnly,
		FinishOnly:     req.FinishOnly,
		MetadataFilter: req.MetadataFilter,
		DBGetFilter: models.DBGetFilter{
			Limit:    req.Limit,
			ID:       req.Search,
			Keyword:  req.Search,
			SortDesc: req.OrderDesc,
		},
	}

	// Enable pagination if Page and Limit called
	if req.Page > 0 && req.Limit > 0 {
		filter.Offset = (req.Page - 1) * req.Limit
	}

	// If user only needs to return number
	if req.NumberOnly {
		jobCount, err := job.Count(filter)
		if err != nil {
			return &DaemonError{"Failed when counting the job", err}
		}

		if err := jc.SendPayload(payload.JobResponseCount{Count: jobCount}); err != nil {
			return &DaemonError{"Invalid metadata: Failed to send payload", err}
		}

		return nil
	}

	rawJobs, err := job.Get(filter)
	if err != nil {
		return &DaemonError{"Failed when fetch the job", err}
	}
	jobs, err := job.BulkToPayload(rawJobs)
	if err != nil {
		return &DaemonError{"Failed when transforming raw job", err}
	}

	if err := jc.SendPayload(jobs); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

	return nil
}

// WaitJob handles requests to wait for a specific job to complete.
// It continuously checks for the absence of the job's lock file and
// responds to the client with the job's final status once completed.
// It also handles client connection timeouts.
func (d *DaemonStruct) WaitJob(jc *JobContext) error {
	var req payload.JobSearchMetadata
	if err := jc.Payload.UnmarshalMetadata(&req); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	jobModel, err := models.NewJobModel(jc.daemon.DB, payload.JobResponse{})
	if err != nil {
		return &DaemonError{"Failed when initialize db model", err}
	}

	filter := &models.JobFilter{
		ActiveOnly: true,
		DBGetFilter: models.DBGetFilter{
			ID:       req.Search,
			Keyword:  req.Search,
			Limit:    1,
			SortDesc: true,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	rawJob, err := jobModel.WaitJob(ctx, cancel, filter)
	if err != nil {
		return &DaemonError{"Failed when waiting the job", err}
	}
	job, err := rawJob.ToPayload()
	if err != nil {
		return &DaemonError{"Failed when transforming job", err}
	}

	if err := jc.SendPayload(job); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

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
	var req payload.JobSearchMetadata
	if err := jc.Payload.UnmarshalMetadata(&req); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	jobModel, err := models.NewJobModel(jc.daemon.DB, payload.JobResponse{})
	if err != nil {
		return &DaemonError{"Failed when initialize db model", err}
	}

	filter := &models.JobFilter{
		DBGetFilter: models.DBGetFilter{
			ID:       req.Search,
			Keyword:  req.Search,
			Limit:    1,
			SortDesc: true,
		},
	}
	jobs, err := jobModel.Get(filter)
	if err != nil {
		return &DaemonError{"Failed when finding job", err}
	}
	if sizeJob := len(jobs); sizeJob < 1 {
		return &DaemonError{"Job not found", fmt.Errorf("len: %v", sizeJob)}
	}

	jobResp, err := jobs[0].ToPayload()
	if err != nil {
		return &DaemonError{"Failed when parsing the payload", err}
	}

	if err := jc.SendPayload(jobResp); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

	return nil
}

// StopJob handles requests to stop the specific job.
// If the job exist, the return is JobResponse. If the job not exist, the return is an empty JobResponse.
func (d *DaemonStruct) StopJob(jc *JobContext) error {
	var req payload.JobSearchMetadata
	if err := jc.Payload.UnmarshalMetadata(&req); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata", err}
	}

	jobModel, err := models.NewJobModel(jc.daemon.DB, payload.JobResponse{})
	if err != nil {
		return &DaemonError{"Failed when initialize db model", err}
	}

	filter := &models.JobFilter{
		DBGetFilter: models.DBGetFilter{
			ID:       req.Search,
			Keyword:  req.Search,
			Limit:    1,
			SortDesc: true,
		},
	}
	jobs, err := jobModel.Get(filter)
	if err != nil {
		return &DaemonError{"Failed when finding job", err}
	}
	if sizeJob := len(jobs); sizeJob < 1 {
		return &DaemonError{"Job not found", fmt.Errorf("len: %v", sizeJob)}
	}

	job := jobs[0]
	if err := syscall.Kill(-job.PID, syscall.SIGTERM); err != nil {
		return err
	}

	job.Status = int(payload.JOB_STOPPED)
	if err := job.Update(); err != nil {
		log.Printf("[WARNING] Failed when updating status: %+v", err)
	}

	jobPayload, err := job.ToPayload()
	if err != nil {
		return &DaemonError{"Failed when parsing the payload", err}
	}
	if err := jc.SendPayload(jobPayload); err != nil {
		return &DaemonError{"Invalid metadata: Failed to send payload", err}
	}

	return nil
}
