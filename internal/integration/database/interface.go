package database

import "io"

// DumpParams represents parameters for database dump operations
// Different database types can have their own specific parameters
type DumpParams interface{}

// DatabaseClient is the interface that all database clients must implement
type DatabaseClient interface {
	// Test tests the connection to the database
	Test(version string, connString string) error

	// DumpZip creates a compressed backup of the database and returns it as an io.Reader
	// The backup format is ZIP containing the dump file(s)
	DumpZip(version string, connString string, params DumpParams) io.Reader

	// RestoreZip restores a database from a ZIP backup file
	// isLocal indicates whether the zip file is local (true) or a URL (false)
	RestoreZip(version string, connString string, isLocal bool, zipURLOrPath string) error

	// ParseVersion validates and parses the version string for the database type
	ParseVersion(version string) (interface{}, error)

	// GetSupportedVersions returns a list of supported versions for this database type
	GetSupportedVersions() []string

	// GetDatabaseType returns the database type identifier (e.g., "postgresql", "clickhouse")
	GetDatabaseType() string
}
