package dblib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

const sqliteMinJSONVersion = 3038000 // SQLite 3.38.0

// CheckSQLiteJSONFunctions checks if the connected SQLite database supports JSON functions (version >= 3.38.0).
func CheckSQLiteJSONFunctions(db *sqlx.DB) (bool, error) {
	var version string
	err := db.Get(&version, "SELECT sqlite_version()")
	if err != nil {
		return false, fmt.Errorf("failed to get SQLite version: %w", err)
	}

	// Parse the version string (e.g., "3.38.5")
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return false, fmt.Errorf("unrecognized SQLite version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, fmt.Errorf("failed to parse major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, fmt.Errorf("failed to parse minor version: %w", err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return false, fmt.Errorf("failed to parse patch version: %w", err)
	}

	// Convert to a comparable integer format (e.g., 3.38.0 -> 3038000)
	currentVersionInt := major*1000000 + minor*1000 + patch

	return currentVersionInt >= sqliteMinJSONVersion, nil
}

// SanitizeJsonKey will sanitize the json key.
func SanitizeJsonKey(key string) string {
	jsonKeySanitizer := regexp.MustCompile(`[^a-zA-Z0-9_]`)

	if key == "" {
		return ""
	}
	return jsonKeySanitizer.ReplaceAllString(key, "")
}
