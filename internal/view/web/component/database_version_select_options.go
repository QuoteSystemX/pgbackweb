package component

import (
	"database/sql"

	"github.com/eduardolat/pgbackweb/internal/integration/clickhouse"
	"github.com/eduardolat/pgbackweb/internal/integration/database"
	"github.com/eduardolat/pgbackweb/internal/integration/postgres"
	nodx "github.com/nodxdev/nodxgo"
)

func DatabaseVersionSelectOptions(dbType string, selectedVersion sql.NullString) nodx.Node {
	var versions []string
	var versionPrefix string

	switch dbType {
	case database.DatabaseTypePostgreSQL:
		pgVersions := postgres.PGVersionsDesc
		versions = make([]string, len(pgVersions))
		for i, v := range pgVersions {
			versions[i] = v.Value.Version
		}
		versionPrefix = "PostgreSQL "
	case database.DatabaseTypeClickHouse:
		chClient := clickhouse.New()
		versions = chClient.GetSupportedVersions()
		versionPrefix = "ClickHouse "
	default:
		// Default to PostgreSQL if unknown type
		pgVersions := postgres.PGVersionsDesc
		versions = make([]string, len(pgVersions))
		for i, v := range pgVersions {
			versions[i] = v.Value.Version
		}
		versionPrefix = "PostgreSQL "
	}

	return nodx.Map(
		versions,
		func(version string) nodx.Node {
			return nodx.Option(
				nodx.Value(version),
				nodx.Textf("%s%s", versionPrefix, version),
				nodx.If(
					selectedVersion.Valid && selectedVersion.String == version,
					nodx.Selected(""),
				),
			)
		},
	)
}
