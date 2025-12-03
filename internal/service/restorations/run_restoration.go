package restorations

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
	"github.com/eduardolat/pgbackweb/internal/logger"
	"github.com/google/uuid"
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

// RunRestoration runs a backup restoration
func (s *Service) RunRestoration(
	ctx context.Context,
	executionID uuid.UUID,
	databaseID uuid.NullUUID,
	connString string,
) error {
	updateRes := func(params dbgen.RestorationsServiceUpdateRestorationParams) error {
		_, err := s.dbgen.RestorationsServiceUpdateRestoration(
			ctx, params,
		)
		return err
	}

	logError := func(err error) {
		dbID := "empty"
		if databaseID.Valid {
			dbID = databaseID.UUID.String()
		}
		logger.Error("error running restoration", logger.KV{
			"execution_id": executionID.String(),
			"database_id":  dbID,
			"error":        err.Error(),
		})
	}

	res, err := s.CreateRestoration(ctx, dbgen.RestorationsServiceCreateRestorationParams{
		ExecutionID: executionID,
		DatabaseID:  databaseID,
		Status:      "running",
	})
	if err != nil {
		logError(err)
		return err
	}

	if !databaseID.Valid && connString == "" {
		err := fmt.Errorf("database_id or connection_string must be provided")
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	execution, err := s.executionsService.GetExecution(ctx, executionID)
	if err != nil {
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	if execution.Status != "success" || !execution.Path.Valid {
		err := fmt.Errorf("backup execution must be successful")
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	if databaseID.Valid {
		db, err := s.databasesService.GetDatabase(ctx, databaseID.UUID)
		if err != nil {
			logError(err)
			return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
				ID:         res.ID,
				Status:     sql.NullString{Valid: true, String: "failed"},
				Message:    sql.NullString{Valid: true, String: err.Error()},
				FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
			})
		}
		connString = db.DecryptedConnectionString
	}

	// Get database client based on database type
	dbClient, err := s.ints.GetDatabaseClient(execution.DatabaseDatabaseType)
	if err != nil {
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	// Extract version string (handles both string and sql.NullString after SQLC regeneration)
	databaseVersion := extractVersionString(execution.DatabaseVersion)

	// Test database connection
	err = dbClient.Test(databaseVersion, connString)
	if err != nil {
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	isLocal, zipURLOrPath, err := s.executionsService.GetExecutionDownloadLinkOrPath(
		ctx, executionID,
	)
	if err != nil {
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	err = dbClient.RestoreZip(databaseVersion, connString, isLocal, zipURLOrPath)
	if err != nil {
		logError(err)
		return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
			ID:         res.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	logger.Info("backup restored successfully", logger.KV{
		"restoration_id": res.ID.String(),
		"execution_id":   executionID.String(),
	})
	return updateRes(dbgen.RestorationsServiceUpdateRestorationParams{
		ID:         res.ID,
		Status:     sql.NullString{Valid: true, String: "success"},
		Message:    sql.NullString{Valid: true, String: "Backup restored successfully"},
		FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
	})
}
