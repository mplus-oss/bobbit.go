package metadata

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// mustExecQuery act like Stored Procedure every single Bobbit
// connected to SQLite.
func mustExecQuery(db *sqlx.DB) error {
	trx := db.MustBegin()
	qExec := func (purpose, query string, args ...any) {
		log.Printf("MustExec: %v", purpose)
		trx.MustExec(query, args...)
	}

	// Start writing from here
	qExec(
		"Normalize Metadata",
		"UPDATE jobs SET metadata='{}' WHERE metadata=''",
	)

	return trx.Commit()
}
