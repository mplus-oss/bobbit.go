package payload

import (
	"encoding/json"
	"time"
)

type PayloadRequestEnum int32

const (
	REQUEST_EXECUTE_JOB PayloadRequestEnum = 1 << iota
	REQUEST_LIST
	REQUEST_WAIT
	REQUEST_STATUS
	REQUEST_VIBE_CHECK
)

type PayloadRegularMetadata map[string]any
type JobPayload struct {
	Request   PayloadRequestEnum     `json:"request"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  PayloadRegularMetadata `json:"metadata,omitempty"`
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
