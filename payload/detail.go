package payload

import "time"

// JobDetailMetadata provides detailed information about a specific job.
//
// This struct can be used to create Job with REQUEST_EXECUTE_JOB payload request.
type JobDetailMetadata struct {
	// ID is the unique identifier for the job.
	ID string `json:"id"`

	// JobName is the name given to the job.
	JobName string `json:"job_name"`

	// Deprecated: See CreatedAt and UpdatedAt
	//
	// Timestamp indicates when the job was recorded.
	Timestamp time.Time `json:"timestamp"`

	// Command specifies the command-line arguments used to execute the job. Required for creating new job.
	Command []string `json:"command,omitempty"`

	// Metadata contains additional regular payload metadata associated with the job.
	Metadata PayloadRegularMetadata `json:"metadata,omitempty"`

	// CreatedAt indicates when the job was recorded first time.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt indicates every single changes job status.
	UpdatedAt time.Time `json:"updated_at"`
}
