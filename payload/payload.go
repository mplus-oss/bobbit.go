package payload

import (
	"encoding/json"
	"time"
)

// PayloadRequestEnum defines the type of request being made in a job payload.
type PayloadRequestEnum int32

const (
	// REQUEST_EXECUTE_JOB indicates a request to execute a new job. Return of this request is void.
	REQUEST_EXECUTE_JOB PayloadRequestEnum = 1 << iota
	// REQUEST_LIST indicates a request to list existing jobs. Return of this request is []JobResponse.
	REQUEST_LIST
	// REQUEST_WAIT indicates a request to wait for a job to complete. Return of this request is JobResponse.
	REQUEST_WAIT
	// REQUEST_STATUS indicates a request to get the status of a specific job. Return of this request is JobResponse.
	REQUEST_STATUS
	// REQUEST_VIBE_CHECK indicates a request to perform a health or liveness check. Return of this request is void.
	REQUEST_VIBE_CHECK
	// REQUEST_STOP indicates a request to stop a job.
	// If the job exist, the return is JobResponse. If the job not exist, the return is an empty JobResponse.
	REQUEST_STOP
)

// PayloadRegularMetadata is a type alias for any, used to hold arbitrary metadata
// associated with a job request. This allows for flexible and extensible payload structures.
type PayloadRegularMetadata any

// JobPayload represents the standard structure for job-related requests.
//
// It encapsulates the type of request, a timestamp, and generic metadata.
type JobPayload struct {
	// Request specifies the type of operation to be performed (e.g., execute, list, status).
	Request PayloadRequestEnum `json:"request"`

	// Timestamp records the time when the request was initiated or last updated.
	Timestamp time.Time `json:"timestamp"`

	// Metadata holds additional, request-specific data. Its structure varies based on the 'Request' type.
	Metadata PayloadRegularMetadata `json:"metadata,omitempty"`
}

// UnmarshalMetadata unmarshals the generic 'Metadata' field of the JobPayload
// into a target Go struct or type. It allows strong-typing of the metadata content.
func (j *JobPayload) UnmarshalMetadata(target any) error {
	metadataBytes, err := json.Marshal(j.Metadata)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(metadataBytes, target); err != nil {
		return err
	}

	return nil
}

// MarshalMetadata marshals a given Go struct or type into the generic 'Metadata'
// field of the JobPayload. It converts the structured data into an `any` type suitable
// for the `Metadata` field, typically as a `map[string]any`.
func (j *JobPayload) MarshalMetadata(target any) error {
	metaJSON, err := json.Marshal(target)
	if err != nil {
		return err
	}

	var metaMap map[string]any
	if err := json.Unmarshal(metaJSON, &metaMap); err != nil {
		return err
	}

	j.Metadata = metaMap
	return nil
}
