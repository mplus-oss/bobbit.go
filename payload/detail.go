package payload

import "time"

// JobDetailMetadata provides detailed information about a specific job.
type JobDetailMetadata struct {
	// ID is the unique identifier for the job.
	ID string `json:"id"`

	// JobName is the name given to the job.
	JobName string `json:"job_name"`

	// Timestamp indicates when the job was recorded or last updated.
	Timestamp time.Time `json:"timestamp"`

	// Command specifies the command-line arguments used to execute the job, if applicable.
	Command []string `json:"command,omitempty"`

	// Metadata contains additional regular payload metadata associated with the job.
	Metadata PayloadRegularMetadata `json:"metadata,omitempty"`
}
