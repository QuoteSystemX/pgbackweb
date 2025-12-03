package clickhouse

import (
	"archive/zip"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eduardolat/pgbackweb/internal/integration/database"
	"github.com/eduardolat/pgbackweb/internal/util/strutil"
)

// DumpParams contains the parameters for the clickhouse-backup command
type DumpParams struct {
	// Tables is a list of specific tables to backup (empty means all tables)
	Tables []string

	// AllDatabases backs up all databases
	AllDatabases bool

	// Compression level (0-9, where 0 means no compression)
	Compression int
}

// GetDatabaseType returns the database type identifier
func (Client) GetDatabaseType() string {
	return database.DatabaseTypeClickHouse
}

// GetSupportedVersions returns a list of supported ClickHouse versions
func (Client) GetSupportedVersions() []string {
	// Common stable ClickHouse versions
	return []string{"22.8", "23.8", "24.1", "24.3"}
}

// ParseVersion validates and parses the version string for ClickHouse
func (Client) ParseVersion(version string) (interface{}, error) {
	// Validate version format (should be like "22.8", "23.8", etc.)
	supportedVersions := map[string]bool{
		"22.8": true,
		"23.8": true,
		"24.1": true,
		"24.3": true,
	}

	if !supportedVersions[version] {
		return nil, fmt.Errorf("clickhouse version not allowed: %s", version)
	}

	return version, nil
}

// parseConnectionString parses a ClickHouse connection string and returns
// command-line arguments for clickhouse-client.
// Supports URL format (clickhouse://user:password@host:port/database)
func parseConnectionString(connString string) ([]string, error) {
	// If it starts with clickhouse://, parse as URL
	if strings.HasPrefix(connString, "clickhouse://") {
		parsedURL, err := url.Parse(connString)
		if err != nil {
			return nil, fmt.Errorf("error parsing ClickHouse URL: %w", err)
		}

		args := []string{}

		// Extract host
		host := parsedURL.Hostname()
		if host == "" {
			host = "localhost"
		}
		args = append(args, "--host="+host)

		// Extract port
		port := parsedURL.Port()
		if port == "" {
			port = "9000" // Default ClickHouse port
		}
		args = append(args, "--port="+port)

		// Extract user
		user := parsedURL.User.Username()
		if user == "" {
			user = "default"
		}
		args = append(args, "--user="+user)

		// Extract password
		password, hasPassword := parsedURL.User.Password()
		if hasPassword {
			args = append(args, "--password="+password)
		}

		// Extract database (path)
		database := strings.TrimPrefix(parsedURL.Path, "/")
		if database == "" {
			database = "default"
		}
		args = append(args, "--database="+database)

		return args, nil
	}

	// Otherwise, treat as already-formatted command-line arguments
	// Split by spaces to get individual arguments
	// This maintains backward compatibility with flag-based format
	parts := strings.Fields(connString)
	return parts, nil
}

// Test tests the connection to the ClickHouse database
func (Client) Test(version string, connString string) error {
	// Parse connection string to get command-line arguments
	connArgs, err := parseConnectionString(connString)
	if err != nil {
		return fmt.Errorf("error parsing connection string: %w", err)
	}

	// Build command with connection arguments and query
	args := append(connArgs, "--query", "SELECT 1")
	cmd := exec.Command("clickhouse-client", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"error running clickhouse-client test v%s: %s",
			version, output,
		)
	}

	return nil
}

// DumpZip creates a backup using clickhouse-backup and returns it as a ZIP-compressed io.Reader
func (c *Client) DumpZip(version string, connString string, params database.DumpParams) io.Reader {
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()

		// Create temporary directory for backup
		workDir, err := os.MkdirTemp("", "ch-backup-*")
		if err != nil {
			writer.CloseWithError(fmt.Errorf("error creating temp dir: %w", err))
			return
		}
		defer os.RemoveAll(workDir)

		backupName := fmt.Sprintf("backup-%s", version)
		backupPath := filepath.Join(workDir, backupName)

		// Build clickhouse-backup command
		args := []string{"create", backupName}

		var dumpParams DumpParams
		if chParams, ok := params.(DumpParams); ok {
			dumpParams = chParams
		}

		// Add table-specific options if provided
		if len(dumpParams.Tables) > 0 {
			for _, table := range dumpParams.Tables {
				args = append(args, "--table", table)
			}
		}

		if dumpParams.AllDatabases {
			args = append(args, "--all-databases")
		}

		if dumpParams.Compression > 0 {
			args = append(args, "--compression", fmt.Sprintf("%d", dumpParams.Compression))
		}

		// Run clickhouse-backup create
		cmd := exec.Command("clickhouse-backup", args...)
		cmd.Dir = workDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			writer.CloseWithError(fmt.Errorf(
				"error running clickhouse-backup create v%s: %s",
				version, output,
			))
			return
		}

		// Check if backup directory exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			writer.CloseWithError(fmt.Errorf("backup directory not found: %s", backupPath))
			return
		}

		// Create ZIP archive from backup directory
		zipWriter := zip.NewWriter(writer)
		defer zipWriter.Close()

		// Walk through backup directory and add all files to ZIP
		err = filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Get relative path from backup directory
			relPath, err := filepath.Rel(backupPath, path)
			if err != nil {
				return err
			}

			// Create file in ZIP
			fileWriter, err := zipWriter.Create(relPath)
			if err != nil {
				return err
			}

			// Read and write file content
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(fileWriter, file)
			return err
		})

		if err != nil {
			writer.CloseWithError(fmt.Errorf("error creating zip from backup: %w", err))
			return
		}
	}()

	return reader
}

// RestoreZip restores a ClickHouse database from a ZIP backup file
func (Client) RestoreZip(version string, connString string, isLocal bool, zipURLOrPath string) error {
	workDir, err := os.MkdirTemp("", "ch-restore-*")
	if err != nil {
		return fmt.Errorf("error creating temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	zipPath := strutil.CreatePath(true, workDir, "backup.zip")
	backupPath := strutil.CreatePath(true, workDir, "backup")

	// Download or copy ZIP file
	if isLocal {
		cmd := exec.Command("cp", zipURLOrPath, zipPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error copying ZIP file to temp dir: %s", output)
		}
	} else {
		cmd := exec.Command("wget", "--no-verbose", "-O", zipPath, zipURLOrPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error downloading ZIP file: %s", output)
		}
	}

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return fmt.Errorf("zip file not found: %s", zipPath)
	}

	// Extract ZIP file
	cmd := exec.Command("unzip", "-o", zipPath, "-d", backupPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error unzipping ZIP file: %s", output)
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup directory not found in ZIP file: %s", zipPath)
	}

	// Get backup name from directory (should be the first directory in backupPath)
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("error reading backup directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("backup directory is empty: %s", backupPath)
	}

	// Find first directory entry
	var backupName string
	for _, entry := range entries {
		if entry.IsDir() {
			backupName = entry.Name()
			break
		}
	}

	if backupName == "" {
		// If no directory found, use the files directly
		// This handles the case where backup files are in the root
		backupName = "backup"
	}

	// Run clickhouse-backup restore
	restorePath := filepath.Join(backupPath, backupName)
	cmd = exec.Command("clickhouse-backup", "restore", restorePath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"error running clickhouse-backup restore v%s: %s",
			version, output,
		)
	}

	return nil
}

type Client struct{}

func New() *Client {
	return &Client{}
}
