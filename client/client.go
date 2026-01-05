package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"

	"github.com/mplus-oss/bobbit.go/config"
	"github.com/mplus-oss/bobbit.go/payload"
)

// DaemonConnectionStruct holds the network connection and configuration for a Bobbit daemon client.
type DaemonConnectionStruct struct {
	Connection net.Conn
	config.BobbitClientConfig
}

// New creates and returns a new DaemonConnectionStruct initialized with the provided BobbitConfig.
func New(c config.BobbitClientConfig) *DaemonConnectionStruct {
	return &DaemonConnectionStruct{
		BobbitClientConfig: c,
	}
}

// BuildPayload establishes a connection to the daemon's Unix socket and marshals the provided metadata
// into the JobPayload. It returns an error if the connection fails or metadata marshalling fails.
func (d *DaemonConnectionStruct) BuildPayload(p *payload.JobPayload, metadata any) (err error) {
	conn, err := net.Dial("unix", d.BobbitConfig.SocketPath)
	if err != nil {
		return err
	}
	d.Connection = conn

	if err := p.MarshalMetadata(metadata); err != nil {
		d.Connection.Close()
		return err
	}

	return nil
}

// SendPayload encodes and sends the given target object over the established daemon connection.
// It returns an error if encoding or writing fails.
func (d *DaemonConnectionStruct) SendPayload(target any) error {
	if err := json.NewEncoder(d.Connection).Encode(target); err != nil {
		return err
	}
	return nil
}

// GetPayload decodes the response from the daemon connection into the provided target object.
// It returns an error if decoding fails.
func (d *DaemonConnectionStruct) GetPayload(target any) error {
	var raw json.RawMessage
	if err := json.NewDecoder(d.Connection).Decode(&raw); err != nil {
		return err
	}

	var errorPayload payload.JobErrorResponse
	if err := json.Unmarshal(raw, &errorPayload); err == nil && errorPayload.Error != "" {
		return errors.New(errorPayload.Error)
	}

	if err := json.Unmarshal(raw, target); err != nil {
		return err
	}

	return nil
}

// TestConnection verifies if the daemon is reachable and responding.
// It sends a simple vibe check payload and returns nil if successful.
func (d *DaemonConnectionStruct) TestConnection() error {
	p := payload.JobPayload{Request: payload.REQUEST_VIBE_CHECK}
	if err := d.BuildPayload(&p, make(map[string]string, 1)); err != nil {
		return err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return err
	}

	return nil
}

// Status retrieves the detailed status of a specific job by its ID.
// Returns the JobResponse containing job details or an error if the request fails.
func (d *DaemonConnectionStruct) Status(id string) (payload.JobResponse, error) {
	p := payload.JobPayload{Request: payload.REQUEST_STATUS}
	if err := d.BuildPayload(&p, payload.JobSearchMetadata{Search: id}); err != nil {
		return payload.JobResponse{}, err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return payload.JobResponse{}, err
	}

	var job payload.JobResponse
	if err := d.GetPayload(&job); err != nil {
		return payload.JobResponse{}, err
	}

	return job, nil
}

// Create submits a new job execution request to the daemon.
// Takes JobDetailMetadata containing command, name, and other options.
func (d *DaemonConnectionStruct) Create(req payload.JobDetailMetadata) error {
	p := payload.JobPayload{Request: payload.REQUEST_EXECUTE_JOB}
	if err := d.BuildPayload(&p, req); err != nil {
		return err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return err
	}

	return nil
}

// Wait blocks until the specified job has finished execution.
// Returns the final JobResponse or an error if the wait fails.
func (d *DaemonConnectionStruct) Wait(id string) (job payload.JobResponse, err error) {
	p := payload.JobPayload{Request: payload.REQUEST_WAIT}
	if err = d.BuildPayload(&p, payload.JobSearchMetadata{Search: id}); err != nil {
		return job, err
	}
	defer d.Connection.Close()

	if err = d.SendPayload(p); err != nil {
		return job, err
	}

	if err = d.GetPayload(&job); err != nil {
		return job, err
	}

	return job, nil
}

// List retrieves a list of jobs based on the provided search criteria.
func (d *DaemonConnectionStruct) List(req payload.JobSearchMetadata) ([]payload.JobResponse, error) {
	p := payload.JobPayload{Request: payload.REQUEST_LIST}
	if err := d.BuildPayload(&p, req); err != nil {
		return []payload.JobResponse{}, err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return []payload.JobResponse{}, err
	}

	var jobs []payload.JobResponse
	if err := d.GetPayload(&jobs); err != nil {
		return []payload.JobResponse{}, err
	}

	return jobs, nil
}

// ListCount returns the number of jobs matching the activeOnly criteria.
// If activeOnly is true, counts active jobs; otherwise counts finished/failed jobs.
// Note: This matches the user's specific logic for active vs finish only.
func (d *DaemonConnectionStruct) ListCount(activeOnly bool) (int, error) {
	p := payload.JobPayload{Request: payload.REQUEST_LIST}
	req := payload.JobSearchMetadata{NumberOnly: true}

	if activeOnly {
		req.ActiveOnly = true
	} else {
		req.FinishOnly = true
	}

	if err := d.BuildPayload(&p, req); err != nil {
		return 0, err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return 0, err
	}

	var count payload.JobResponseCount
	if err := d.GetPayload(&count); err != nil {
		return 0, err
	}

	return count.Count, nil
}

// Stop sends a request to stop a running job by its name or ID.
// Returns the JobResponse of the stopped job or an error if the request fails.
func (d *DaemonConnectionStruct) Stop(jobNameOrId string) (payload.JobResponse, error) {
	p := payload.JobPayload{Request: payload.REQUEST_STOP}
	req := payload.JobSearchMetadata{
		Search: jobNameOrId,
	}
	if err := d.BuildPayload(&p, req); err != nil {
		return payload.JobResponse{}, err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return payload.JobResponse{}, err
	}

	var job payload.JobResponse
	if err := d.GetPayload(&job); err != nil {
		return payload.JobResponse{}, err
	}

	return job, nil
}

// FindJob attempts to locate a single job matching the provided query parameters.
// This uses REQUEST_STATUS under the hood, similar to Status but with a full metadata struct.
func (d *DaemonConnectionStruct) FindJob(query payload.JobSearchMetadata) (payload.JobResponse, error) {
	p := payload.JobPayload{Request: payload.REQUEST_STATUS}
	if err := d.BuildPayload(&p, query); err != nil {
		return payload.JobResponse{}, err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return payload.JobResponse{}, err
	}

	var job payload.JobResponse
	if err := d.GetPayload(&job); err != nil {
		return payload.JobResponse{}, err
	}

	return job, nil
}

// TailJobLogWithContext streams a job's log file in real-time with context support.
// It takes a context for cancellation, job ID or search string, and a callback function.
// The callback receives the log line as a string. If the callback returns an error, streaming stops.
// Returns an error if the request fails, job is not found, or context is cancelled.
func (d *DaemonConnectionStruct) TailJobLogWithContext(ctx context.Context, jobIDOrName string, follow bool, onLine func(string) error) error {
	p := payload.JobPayload{Request: payload.REQUEST_TAIL_LOG}
	search := payload.JobSearchMetadata{Search: jobIDOrName, Follow: follow}
	if err := d.BuildPayload(&p, search); err != nil {
		return err
	}
	defer d.Connection.Close()

	if err := d.SendPayload(p); err != nil {
		return err
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			d.Connection.Close()
		case <-done:
		}
	}()

	decoder := json.NewDecoder(d.Connection)
	// Scan every single new request.
	// Surely will cannot bumped to another connection...
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var line map[string]string
		if err := decoder.Decode(&line); err != nil {
			if err == io.EOF {
				return nil
			}
			// Check if error is due to context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		if logLine, ok := line["line"]; ok {
			if err := onLine(logLine); err != nil {
				return err
			}
		}
	}
}
