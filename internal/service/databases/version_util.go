package databases

import (
	"database/sql"
	"reflect"
)

// extractVersionString extracts version string from either string or sql.NullString
// This handles both cases: before SQLC regeneration (string) and after (sql.NullString)
func extractVersionString(version interface{}) string {
	if version == nil {
		return ""
	}

	// Handle sql.NullString
	if ns, ok := version.(sql.NullString); ok {
		if ns.Valid {
			return ns.String
		}
		return ""
	}

	// Handle string
	if s, ok := version.(string); ok {
		return s
	}

	// Handle via reflection for other types
	v := reflect.ValueOf(version)
	if v.Kind() == reflect.String {
		return v.String()
	}

	return ""
}

