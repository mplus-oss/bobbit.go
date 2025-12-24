package models

import "github.com/jmoiron/sqlx"

type BaseModel struct {
	DB *sqlx.DB `db:"-"`
}

// DBGetFilter defines the criteria used to query and filter job records from the database.
// It is typically passed to repository functions to refine the result set.
type DBGetFilter struct {
	// ID specifies a unique job identifier to filter by.
	// If provided, the query typically returns a single match.
	ID string

	// Keyword specifies a search term for partial matching.
	// This is usually applied against the job name or ID (e.g., "backup" matches "db-backup-01").
	Keyword string

	// Limit restricts the maximum number of records returned by the query.
	// A value of 0 typically implies no limit or uses the system default.
	Limit int

	// SortDesc determines the sorting order of the results based on the timestamp.
	// If true, results are ordered descending (newest first). If false, they are ordered ascending (oldest first).
	SortDesc bool

	// Offset specifies the number of records to skip before starting to return results.
	// Used in conjunction with Limit for pagination.
	Offset int
}
