package payload

import "time"

type JobRequestMetadata struct {
	ID        string                 `json:"id"`
	JobName   string                 `json:"job_name"`
	Timestamp time.Time              `json:"timestamp"`
	Command   []string               `json:"command,omitempty"`
	Metadata  PayloadRegularMetadata `json:"metadata,omitempty"`
	JobPayload
}
