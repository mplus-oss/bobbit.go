package payload

// JobError is used to unmarshal error messages from a JSON response payload.
// It expects a JSON object with a single "error" key containing the error message.
type JobErrorResponse struct {
	Error string `json:"error"`
}
