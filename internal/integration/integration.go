package integration

import (
	"fmt"

	"github.com/eduardolat/pgbackweb/internal/integration/clickhouse"
	"github.com/eduardolat/pgbackweb/internal/integration/database"
	"github.com/eduardolat/pgbackweb/internal/integration/postgres"
	"github.com/eduardolat/pgbackweb/internal/integration/storage"
)

type Integration struct {
	DatabaseClients map[string]database.DatabaseClient
	StorageClient   *storage.Client
	// PGClient is kept for backward compatibility
	PGClient *postgres.Client
}

func New() *Integration {
	pgClient := postgres.New()
	chClient := clickhouse.New()
	storageClient := storage.New()

	dbClients := make(map[string]database.DatabaseClient)
	dbClients[database.DatabaseTypePostgreSQL] = pgClient
	dbClients[database.DatabaseTypeClickHouse] = chClient

	return &Integration{
		DatabaseClients: dbClients,
		StorageClient:   storageClient,
		PGClient:        pgClient, // Keep for backward compatibility
	}
}

// GetDatabaseClient returns the database client for the given database type
func (i *Integration) GetDatabaseClient(dbType string) (database.DatabaseClient, error) {
	client, ok := i.DatabaseClients[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return client, nil
}
