package payload

import (
	"encoding/json"
	"time"
)

type PayloadRequestEnum int32

const (
	EXECUTE_JOB PayloadRequestEnum = 1 << iota
	LIST
	WAIT
)

type JobPayload struct {
	ID        string             `json:"id"`
	Request   PayloadRequestEnum `json:"request"`
	Command   []string           `json:"command"`
	Timestamp time.Time          `json:"timestamp"`
	Metadata  map[string]any     `json:"metadata,omitempty"`
}

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
