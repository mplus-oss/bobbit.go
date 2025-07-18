package payload

type JobStatusEnum int32

const (
	JOB_RUNNING JobStatusEnum = 1 << iota
	JOB_FINISH
	JOB_FAILED
	JOB_NOT_RUNNING
)

type JobResponse struct {
	Status   JobStatusEnum `json:"status"`
	ExitCode int           `json:"exitcode"`
	JobDetailMetadata
}
