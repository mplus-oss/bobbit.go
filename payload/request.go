package payload

import "time"

type JobRequestMetadata struct {
	ID        string         `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Command   []string       `json:"command,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	JobPayload
}
