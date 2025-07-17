package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/mplus-oss/bobbit.go/internal/lib"
	"github.com/mplus-oss/bobbit.go/payload"
)

func (d *DaemonStruct) HandleVibeCheck(jc *JobContext) error {
	var payload payload.PayloadRegularMetadata
	if err := jc.Payload.UnmarshalMetadata(&payload); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata.", err}
	}
	return nil
}

func (d *DaemonStruct) HandleJob(jc *JobContext) error {
	var payload payload.JobRequestMetadata
	if err := jc.Payload.UnmarshalMetadata(&payload); err != nil {
		return &DaemonError{"Invalid metadata: Failed to unmarshal request metadata: %v", err}
	}

	if payload.JobName == "" || len(payload.Command) < 1 {
		return &DaemonError{"Invalid payload: JobName or Command not provided.", nil}
	}
	if payload.ID == "" {
		hash, err := lib.GenerateRandomHash(16)
		if err != nil {
			return &DaemonError{"Failed to create Hash for job: %v", err}
		}
		payload.ID = hash
	}
	if payload.Timestamp.IsZero() {
		payload.Timestamp = jc.Payload.Timestamp
	}

	lockFile := d.GenerateJobDataFilename(payload, DAEMON_LOCKFILE)
	logFile := d.GenerateJobDataFilename(payload, DAEMON_LOGFILE)
	exitCodeFile := d.GenerateJobDataFilename(payload, DAEMON_EXITCODE)
	metadataFile := d.GenerateJobDataFilename(payload, DAEMON_METADATA)

	log.Printf("Entering HandleJob Context: %v", jc)
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
	log.Printf("DONE: %s %s exit=%d", payload.ID, payload.JobName, exitCode)
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

	log.Printf("Entering ListJob Context: %v", jc)
	jobIDs := make(map[string]bool)
	for _, file := range files {
		jobfile := d.SplitFilenameFromExtfile(file.Name())
		jobIDs[jobfile] = true
	}

	allJobs := []payload.JobStatus{}
	for id := range jobIDs {
		metadata, err := d.ParseJobDataFilename(id)
		if err != nil {
			log.Println(err)
			continue
		}
		status := payload.JobStatus{JobRequestMetadata: metadata}
		status.Status = d.ParseExitCode(metadata)

		if statusRequest.RequestMeta {
			if metaBytes, err := os.ReadFile(d.GenerateJobDataFilename(metadata, DAEMON_METADATA)); err == nil {
				err := json.Unmarshal(metaBytes, &status.Metadata)
				if err != nil {
					log.Printf("Failed to unmarshal metadata from %s job: %v", id, err)
					continue
				}
			}
		}
		if _, err := os.Stat(d.GenerateJobDataFilename(metadata, DAEMON_LOCKFILE)); err == nil {
			status.Status = payload.JOB_RUNNING
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

func (d *DaemonStruct) WaitJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobSearchMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return err
	}

	job, err := d.FindJobDataFilename(statusRequest)
	if err != nil {
		return err
	}
	log.Printf("Got job: %v\n", job)

	if job.ID == "" {
		if err := jc.SendPayload(payload.JobStatus{Status: payload.JOB_NOT_RUNNING}); err != nil {
			return err
		}
	}

	lockFile := d.GenerateJobDataFilename(job, DAEMON_LOCKFILE)
	log.Printf("Waiting job: %v\n", job)
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

	finalStatus := d.ParseExitCode(job)
	if err := jc.SendPayload(payload.JobStatus{Status: finalStatus}); err != nil {
		return err
	}

	log.Printf("Job waiting finish: %v\n", job)
	return nil
}

func (d *DaemonStruct) StatusJob(jc *JobContext) error {
	p := jc.Payload

	var statusRequest payload.JobSearchMetadata
	if err := p.UnmarshalMetadata(&statusRequest); err != nil {
		return err
	}

	job, err := d.FindJobDataFilename(statusRequest)
	if err != nil {
		return err
	}

	if job.ID == "" {
		if err := jc.SendPayload(payload.JobStatus{Status: payload.JOB_NOT_RUNNING}); err != nil {
			return err
		}
	}

	finalStatus := d.ParseExitCode(job)
	if err := jc.SendPayload(payload.JobStatus{Status: finalStatus, JobRequestMetadata: job}); err != nil {
		return err
	}

	return nil
}
