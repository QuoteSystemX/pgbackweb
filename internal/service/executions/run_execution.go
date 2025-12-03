package executions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
	"github.com/eduardolat/pgbackweb/internal/integration/clickhouse"
	"github.com/eduardolat/pgbackweb/internal/integration/database"
	"github.com/eduardolat/pgbackweb/internal/integration/postgres"
	"github.com/eduardolat/pgbackweb/internal/logger"
	"github.com/eduardolat/pgbackweb/internal/util/strutil"
	"github.com/eduardolat/pgbackweb/internal/util/timeutil"
	"github.com/google/uuid"
)

// RunExecution runs a backup execution
func (s *Service) RunExecution(ctx context.Context, backupID uuid.UUID) error {
	updateExec := func(params dbgen.ExecutionsServiceUpdateExecutionParams) error {
		if params.Status.String == "success" {
			s.webhooksService.RunExecutionSuccess(backupID)
		}

		if params.Status.String == "failed" {
			s.webhooksService.RunExecutionFailed(backupID)
		}

		_, err := s.dbgen.ExecutionsServiceUpdateExecution(
			ctx, params,
		)
		return err
	}

	logError := func(err error) {
		logger.Error("error running backup", logger.KV{
			"backup_id": backupID.String(),
			"error":     err.Error(),
		})
	}

	back, err := s.dbgen.ExecutionsServiceGetBackupData(
		ctx, dbgen.ExecutionsServiceGetBackupDataParams{
			BackupID:      backupID,
			EncryptionKey: s.env.PBW_ENCRYPTION_KEY,
		},
	)
	if err != nil {
		logError(err)
		return err
	}

	ex, err := s.CreateExecution(ctx, dbgen.ExecutionsServiceCreateExecutionParams{
		BackupID: backupID,
		Status:   "running",
	})
	if err != nil {
		logError(err)
		return err
	}

	if !back.BackupIsLocal {
		err = s.ints.StorageClient.S3Test(
			back.DecryptedDestinationAccessKey, back.DecryptedDestinationSecretKey,
			back.DestinationRegion.String, back.DestinationEndpoint.String,
			back.DestinationBucketName.String,
		)
		if err != nil {
			logError(err)
			return updateExec(dbgen.ExecutionsServiceUpdateExecutionParams{
				ID:         ex.ID,
				Status:     sql.NullString{Valid: true, String: "failed"},
				Message:    sql.NullString{Valid: true, String: err.Error()},
				FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
			})
		}
	}

	// Get database client based on database type
	dbClient, err := s.ints.GetDatabaseClient(back.DatabaseDatabaseType)
	if err != nil {
		logError(err)
		return updateExec(dbgen.ExecutionsServiceUpdateExecutionParams{
			ID:         ex.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	// Test database connection
	err = dbClient.Test(back.DatabaseVersion, back.DecryptedDatabaseConnectionString)
	if err != nil {
		logError(err)
		return updateExec(dbgen.ExecutionsServiceUpdateExecutionParams{
			ID:         ex.ID,
			Status:     sql.NullString{Valid: true, String: "failed"},
			Message:    sql.NullString{Valid: true, String: err.Error()},
			FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		})
	}

	// Create dump parameters based on database type
	var dumpParams database.DumpParams
	switch back.DatabaseDatabaseType {
	case "postgresql":
		dumpParams = postgres.DumpParams{
			DataOnly:   back.BackupOptDataOnly,
			SchemaOnly: back.BackupOptSchemaOnly,
			Clean:      back.BackupOptClean,
			IfExists:   back.BackupOptIfExists,
			Create:     back.BackupOptCreate,
			NoComments: back.BackupOptNoComments,
		}
	case "clickhouse":
		// ClickHouse backup parameters
		// Note: PostgreSQL-specific options are not applicable to ClickHouse
		// ClickHouse uses different backup mechanisms
		dumpParams = clickhouse.DumpParams{
			Tables:       []string{}, // Empty means all tables
			AllDatabases: false,      // Can be configured per backup if needed
			Compression:  0,          // Default compression
		}
	default:
		// For unknown database types, use nil
		dumpParams = nil
	}

	dumpReader := dbClient.DumpZip(back.DatabaseVersion, back.DecryptedDatabaseConnectionString, dumpParams)

	date := time.Now().Format(timeutil.LayoutSlashYYYYMMDD)
	file := fmt.Sprintf(
		"dump-%s-%s.zip",
		time.Now().Format(timeutil.LayoutYYYYMMDDHHMMSS),
		uuid.NewString(),
	)
	path := strutil.CreatePath(false, back.BackupDestDir, date, file)
	fileSize := int64(0)

	if back.BackupIsLocal {
		fileSize, err = s.ints.StorageClient.LocalUpload(path, dumpReader)
		if err != nil {
			logError(err)
			return updateExec(dbgen.ExecutionsServiceUpdateExecutionParams{
				ID:         ex.ID,
				Status:     sql.NullString{Valid: true, String: "failed"},
				Message:    sql.NullString{Valid: true, String: err.Error()},
				Path:       sql.NullString{Valid: true, String: path},
				FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
			})
		}
	}

	if !back.BackupIsLocal {
		fileSize, err = s.ints.StorageClient.S3Upload(
			back.DecryptedDestinationAccessKey, back.DecryptedDestinationSecretKey,
			back.DestinationRegion.String, back.DestinationEndpoint.String,
			back.DestinationBucketName.String, path, dumpReader,
		)
		if err != nil {
			logError(err)
			return updateExec(dbgen.ExecutionsServiceUpdateExecutionParams{
				ID:         ex.ID,
				Status:     sql.NullString{Valid: true, String: "failed"},
				Message:    sql.NullString{Valid: true, String: err.Error()},
				Path:       sql.NullString{Valid: true, String: path},
				FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
			})
		}
	}

	logger.Info("backup created successfully", logger.KV{
		"backup_id":    backupID.String(),
		"execution_id": ex.ID.String(),
	})
	return updateExec(dbgen.ExecutionsServiceUpdateExecutionParams{
		ID:         ex.ID,
		Status:     sql.NullString{Valid: true, String: "success"},
		Message:    sql.NullString{Valid: true, String: "Backup created successfully"},
		Path:       sql.NullString{Valid: true, String: path},
		FinishedAt: sql.NullTime{Valid: true, Time: time.Now()},
		FileSize:   sql.NullInt64{Valid: true, Int64: fileSize},
	})
}
