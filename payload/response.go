package payload

// JobStatusEnum represents the status of a job.
type JobStatusEnum int32

const (
	// JOB_RUNNING indicates that the job is currently running.
	JOB_RUNNING JobStatusEnum = 1 << iota
	// JOB_FINISH indicates that the job has completed successfully.
	JOB_FINISH
	// JOB_FAILED indicates that the job has failed.
	JOB_FAILED
	// JOB_NOT_RUNNING indicates that the job is not currently running.
	JOB_NOT_RUNNING
)

// JobResponse represents the detailed response for a job query.
//
// It includes the job's status, exit code, and additional job details.
type JobResponse struct {
	// Status indicates the current status of the job.
	Status JobStatusEnum `json:"status"`

	// ExitCode provides the exit code of the job process.
	ExitCode int `json:"exitcode"`

	// JobDetailMetadata embeds additional metadata about the job.
	JobDetailMetadata
}

// JobResponseCount represents a response containing only the count of jobs.
type JobResponseCount struct {

	// Count is the total number of jobs matching the criteria.
	Count int `json:"count"`
}
