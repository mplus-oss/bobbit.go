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

	// Command specifies the command-line arguments used to execute the job. Required for creating new job.
	Command []string `json:"command,omitempty"`

	// Metadata contains additional regular payload metadata associated with the job.
	Metadata PayloadRegularMetadata `json:"metadata,omitempty"`

	// CreatedAt indicates when the job was recorded first time.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt indicates every single changes job status.
	UpdatedAt time.Time `json:"updated_at"`

	// MetadataFilter allows filtering jobs based on their metadata.
	// It's a map where keys are metadata field names and values are the desired values.
	MetadataFilter map[string]string `json:"metadata_filter,omitempty"`
}
