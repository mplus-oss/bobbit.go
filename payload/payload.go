package payload

type PayloadRequestEnum int32

const (
	EXECUTE_JOB PayloadRequestEnum = 1 << iota
	LIST
	WAIT
)

type JobPayload struct {
	ID      string             `json:"id"`
	Request PayloadRequestEnum `json:"request"`
	Command []string           `json:"command"`
}
