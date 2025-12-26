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
	// Cannot be combined with FinishOnly.
	ActiveOnly bool `json:"active_only,omitempty"`

	// FinishOnly filters results to include only finished jobs whether job is success or failed.
	// Cannot be combined with ActiveOnly.
	FinishOnly bool `json:"finish_only,omitempty"`

	// Page specifies the page number for pagination. When set,
	// the Limit field will be used to determine the maximum number of results per page.
	Page int `json:"page,omitempty"`

	// Limit sets the overall maximum number of results to return.
	// When pagination is enabled via the Page field, it determines the maximum number of results per page.
	Limit int `json:"limit,omitempty"`

	// When true, indicates that only the count of matching jobs should be returned.
	NumberOnly bool `json:"number_only,omitempty"`

	// When true, orders the results in descending order.
	OrderDesc bool `json:"desc,omitempty"`

	// MetadataFilter allows filtering jobs based on their metadata.
	// It's a map where keys are metadata field names and values are the desired values.
	MetadataFilter map[string]string `json:"metadata_filter,omitempty"`
}
