package payload

type JobSearchMetadata struct {
	RequestMeta bool   `json:"request_meta"`
	Search      string `json:"search"`
}
