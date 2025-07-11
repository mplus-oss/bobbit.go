package payload

type JobStatusEnum int32
const (
	JOB_RUNNING JobStatusEnum = 1 << iota
	JOB_FINISH
	JOB_FAILED
)

type JobStatus struct {
	Status      JobStatusEnum `json:"status"`
	ExitCode    int           `json:"exitcode"`
	JobPayload
}

type JobStatusMetadata struct {
	RequestMeta bool `json:"request_meta"`
}
