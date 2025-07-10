package config

type JobPayload struct {
	ID      string   `json:"id"`
	Command []string `json:"command"`
}
