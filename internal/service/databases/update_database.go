package databases

import (
	"context"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
)

// UpdateDatabase updates an existing database entry.
func (s *Service) UpdateDatabase(
	ctx context.Context, params dbgen.DatabasesServiceUpdateDatabaseParams,
) (dbgen.Database, error) {
	// Get database type and version - use existing values if not provided
	dbType := params.DatabaseType.String
	if !params.DatabaseType.Valid {
		// Get existing database to get its type
		existing, err := s.GetDatabase(ctx, params.ID)
		if err != nil {
			return dbgen.Database{}, err
		}
		dbType = existing.DatabaseType
	}

	version := params.Version.String
	if !params.Version.Valid {
		// Get existing database to get its version
		existing, err := s.GetDatabase(ctx, params.ID)
		if err != nil {
			return dbgen.Database{}, err
		}
		version = existing.Version
	}

	connString := params.ConnectionString.String
	if !params.ConnectionString.Valid {
		// Get existing database to get its connection string
		existing, err := s.GetDatabase(ctx, params.ID)
		if err != nil {
			return dbgen.Database{}, err
		}
		connString = existing.DecryptedConnectionString
	}

	err := s.TestDatabase(ctx, dbType, version, connString)
	if err != nil {
		return dbgen.Database{}, err
	}

	params.EncryptionKey = s.env.PBW_ENCRYPTION_KEY
	db, err := s.dbgen.DatabasesServiceUpdateDatabase(ctx, params)

	_ = s.TestDatabaseAndStoreResult(ctx, db.ID)

	return db, err
}
