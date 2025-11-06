package component

import (
	"database/sql"

	"github.com/eduardolat/pgbackweb/internal/integration/database"
	nodx "github.com/nodxdev/nodxgo"
)

func DatabaseTypeSelectOptions(selectedType sql.NullString) nodx.Node {
	types := []struct {
		Value string
		Label string
	}{
		{database.DatabaseTypePostgreSQL, "PostgreSQL"},
		{database.DatabaseTypeClickHouse, "ClickHouse"},
	}

	return nodx.Map(
		types,
		func(t struct {
			Value string
			Label string
		}) nodx.Node {
			return nodx.Option(
				nodx.Value(t.Value),
				nodx.Text(t.Label),
				nodx.If(
					selectedType.Valid && selectedType.String == t.Value,
					nodx.Selected(""),
				),
			)
		},
	)
}
