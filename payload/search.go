package payload

// JobSearchMetadata defines the structure for search criteria used to query job information.
//
// The result of a JobSearchMetadata query typically yields a JobResponse.
//
// If NumberOnly is set to true, the query will instead return a JobResponseCount.
type JobSearchMetadata struct {
	// RequestMeta indicates whether metadata related to the request should be included.
	RequestMeta bool `json:"request_meta,omitempty"`

	// Search specifies the search query string.
	Search string `json:"search,omitempty"`

	// ActiveOnly filters results to include only active jobs.
	ActiveOnly bool `json:"active_only,omitempty"`

	// Limit sets the maximum number of results to return.
	Limit int `json:"limit,omitempty"`

	// When true, indicates that only the count of matching jobs should be returned.
	NumberOnly bool `json:"number_only,omitempty"`

	// When true, orders the results in descending order.
	OrderDesc bool `json:"desc,omitempty"`
}
